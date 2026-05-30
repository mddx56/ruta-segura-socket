package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ConnHost      string
	ConnPort      string
	ConnType      string
	WSPort        string
	APIKey        string
	GRPCAddr      string
	RedisHost     string
	RedisPort     string
	RedisPassword string
}

func Load() (*Config, error) {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error al cargar el archivo .env", err.Error())
	}

	cfg := &Config{
		ConnHost:      getEnv("SS_HOST", "localhost"),
		ConnPort:      getEnv("SS_PORT", "5050"),
		ConnType:      "tcp",
		WSPort:        getEnv("WS_PORT", "8081"),
		APIKey:        getEnv("API_KEY", "dd9sad709asyd8y"),
		GRPCAddr:      getEnv("GRPC_ADDR", "localhost:50051"),
		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
