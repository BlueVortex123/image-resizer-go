// cmd/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	httpadapter "example.com/img-resizer/internal/adapters/http"
	"example.com/img-resizer/internal/adapters/queue"
	redisadapter "example.com/img-resizer/internal/adapters/redis"
	configValues "example.com/img-resizer/internal/config"
	"example.com/img-resizer/internal/domain/ports/requests"
	resizerService "example.com/img-resizer/internal/service"
)

func main() {
	configValues := configValues.Load()

	if configValues.App.Env == "production" {
		log.SetFlags(0) // Disable timestamp logging in production
	} else {
		log.SetFlags(log.LstdFlags | log.Lshortfile) // Enable detailed logging in non-production
	}

	//  Initialize Redis connection
	redisCache, err := redisadapter.NewRedisCacheInstance(
		configValues.Redis.Addr,
		configValues.Redis.Password,
		configValues.Redis.DB,
		configValues.Redis.TTL,
	)

	if err != nil {
		if configValues.App.Debug {
			log.Fatalf("Redis init error: %v", err)
		} else {
			log.Println("Redis not available, continuing without cache")
		}
	}

	// Initialize queue client for async processing (optional)
	var queueClient *queue.QueueClient
	var resizerSvc *resizerService.ResizerService

	if len(os.Args) > 1 && os.Args[1] == "worker" {
		// For worker mode, just use the basic service without queue client
		resizerSvc = resizerService.NewResizerService(redisCache, configValues.Redis.TTL, configValues)
		runWorker(configValues, resizerSvc)
		return
	} else {
		// For server mode, initialize with queue support for async endpoints
		queueClient = queue.NewQueueClient(configValues.Redis.Addr)
		resizerSvc = resizerService.NewResizerServiceWithQueue(redisCache, configValues.Redis.TTL, configValues, queueClient)
		
		// Ensure cleanup on exit
		defer func() {
			if queueClient != nil {
				queueClient.Close()
			}
		}()
	}

	// Initialize HTTP server
	serverAdaptor := httpadapter.NewServerInstance(resizerSvc, configValues) // pass service layer as dependency injection

	fmt.Printf("Server running at %s (env=%s)\n", configValues.Server.Port, configValues.App.Env)

	if err := serverAdaptor.Serve(configValues.Server.Port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}

func runWorker(configValues configValues.Config, resizerSvc *resizerService.ResizerService) {
	log.Printf("🔧 Initializing worker with Redis: %s", configValues.Redis.Addr)

	// Create worker configuration
	workerConfig := queue.WorkerConfig{
		RedisAddr:   configValues.Redis.Addr,
		Concurrency: configValues.Worker.Concurrency,
		QueueName:   "images",
		MaxRetry:    configValues.Worker.MaxRetry,
		TaskTimeout: configValues.Worker.TaskTimeout,
	}

	// Create worker with resize handler
	worker := queue.NewWorker(workerConfig, func(ctx context.Context, p queue.ResizeImagePayload) error {
		// Use the resizer service to process the image
		req := requests.RequestGetResize{
			URL:    p.URL,
			Width:  p.Width,
			Height: p.Height,
		}

		log.Printf("📸 Worker processing resize for %s (%dx%d)", p.URL, p.Width, p.Height)
		
		// Process the resize request using the service
		response := resizerSvc.GetResized(ctx, req)
		if response.Error != nil {
			return response.Error
		}

		log.Printf("✅ Worker completed resize: %s (size=%d bytes, type=%s)", 
			p.URL, len(response.Data), response.MimeType)
		return nil
	})

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		<-c
		log.Println("🛑 Received shutdown signal, shutting down worker...")
		worker.Shutdown()
	}()

	if err := worker.Run(); err != nil {
		log.Fatal(err)
	}
}
