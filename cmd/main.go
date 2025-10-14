// cmd/main.go
package main

import (
	"fmt"
	"log"

	httpadapter "example.com/img-resizer/internal/adapters/http"
	redisadapter "example.com/img-resizer/internal/adapters/redis"
	configValues "example.com/img-resizer/internal/config"
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

	// Initiate service layer
	resizerSvc := resizerService.NewResizerService(redisCache, configValues.Redis.TTL, configValues)

	// Initialize HTTP server
	serverAdaptor := httpadapter.NewServerInstance(resizerSvc, configValues) // pass service layer as dependency injection

	fmt.Printf("Server running at %s (env=%s)\n", configValues.Server.Port, configValues.App.Env)

	if err := serverAdaptor.Serve(configValues.Server.Port); err != nil {
		log.Fatalf("Server stopped: %v", err)
	}
}
