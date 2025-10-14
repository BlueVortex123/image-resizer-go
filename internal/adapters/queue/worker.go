package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

type ResizeImageHandler func(ctx context.Context, payload ResizeImagePayload) error

type Worker struct {
	server      *asynq.Server
	mux         *asynq.ServeMux
	concurrency int
}

type WorkerConfig struct {
	RedisAddr   string
	Concurrency int
	QueueName   string
	MaxRetry    int
	TaskTimeout time.Duration
}

func NewWorker(config WorkerConfig, resizeHandler ResizeImageHandler) *Worker {
	// Default values
	if config.Concurrency <= 0 {
		config.Concurrency = 5
	}
	if config.QueueName == "" {
		config.QueueName = "images"
	}
	if config.MaxRetry <= 0 {
		config.MaxRetry = 3
	}
	if config.TaskTimeout <= 0 {
		config.TaskTimeout = 60 * time.Second
	}

	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.RedisAddr},
		asynq.Config{
			Concurrency: config.Concurrency,
			Queues: map[string]int{
				config.QueueName: 10, // Queue priority
			},
			// Add error handling and retry configuration
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Task %s failed: %v", task.Type(), err)
			}),
			// Configure strict priority
			// StrictPriority: false,
		},
	)

	mux := asynq.NewServeMux()

	// Register the resize image handler with improved error handling
	mux.HandleFunc(TypeResizeImage, func(ctx context.Context, task *asynq.Task) error {
		var payload ResizeImagePayload
		if err := json.Unmarshal(task.Payload(), &payload); err != nil {
			log.Printf("Failed to unmarshal task payload: %v", err)
			return fmt.Errorf("invalid payload: %w", err)
		}

		log.Printf("📸 Processing resize task for %s (%dx%d)", payload.URL, payload.Width, payload.Height)

		// Call the handler with timeout context
		ctx, cancel := context.WithTimeout(ctx, config.TaskTimeout)
		defer cancel()

		if err := resizeHandler(ctx, payload); err != nil {
			log.Printf("❌ Resize task failed for %s: %v", payload.URL, err)
			return err
		}

		log.Printf("✅ Resize task completed for %s", payload.URL)
		return nil
	})

	return &Worker{
		server:      srv,
		mux:         mux,
		concurrency: config.Concurrency,
	}
}

func (worker *Worker) Run() error {
	log.Printf("🚀 Starting worker with %d concurrent workers...", worker.concurrency)
	return worker.server.Run(worker.mux)
}

func (worker *Worker) Shutdown() {
	log.Println("🛑 Shutting down worker gracefully...")
	worker.server.Shutdown()
}

func (worker *Worker) Stop() error {
	log.Println("⏹️ Stopping worker...")
	worker.server.Stop()
	return nil
}
