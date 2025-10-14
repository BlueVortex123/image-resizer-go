// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Env   string
		Debug bool
	}
	Redis struct {
		Addr     string
		Password string
		DB       int
		TTL      time.Duration
	}

	HTTP struct {
		Timeout time.Duration
	}

	Server struct {
		Port string
	}
}

func Load() Config {
	// Load .env (ignore error if not found, so we can use environment variables directly)
	_ = godotenv.Load()

	cfg := Config{}

	// App
	cfg.App.Env = strings.ToLower(getEnv("APP_ENV", "development"))
	cfg.App.Debug = getEnvAsBool("APP_DEBUG", true)

	// Redis
	cfg.Redis.Addr = getEnv("REDIS_ADDR", "localhost:6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = getEnvAsInt("REDIS_DB", 0)
	cfg.Redis.TTL = time.Duration(getEnvAsInt("REDIS_TTL_HOURS", 1)) * time.Hour

	// HTTP Client
	cfg.HTTP.Timeout = time.Duration(getEnvAsInt("HTTP_TIMEOUT_SECONDS", 10)) * time.Second

	// Server
	cfg.Server.Port = getEnv("SERVER_PORT", ":8080")

	fmt.Printf("\n%-15s | %-10s\n", "Config Key", "Value")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("%-15s | %-10s\n", "ENV", cfg.App.Env)
	fmt.Printf("%-15s | %-10v\n", "Debug", cfg.App.Debug)
	fmt.Printf("%-15s | %-10s\n", "Redis Addr", cfg.Redis.Addr)
	fmt.Printf("%-15s | %-10s\n", "Redis Pass", cfg.Redis.Password)
	fmt.Printf("%-15s | %-10d\n", "Redis DB", cfg.Redis.DB)
	fmt.Printf("%-15s | %-10v\n", "Redis TTL", cfg.Redis.TTL)
	fmt.Printf("%-15s | %-10v\n", "HTTP Timeout", cfg.HTTP.Timeout)
	fmt.Printf("%-15s | %-10s\n", "Server Port", cfg.Server.Port)
	fmt.Println()

	return cfg
}

// --- helpers ---

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if valStr, ok := os.LookupEnv(key); ok {
		valStr = strings.ToLower(valStr)
		switch valStr {
		case "1", "true", "yes", "on", "da":
			return true
		}
	}
	return defaultVal
}
