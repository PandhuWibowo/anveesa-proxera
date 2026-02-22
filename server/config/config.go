package config

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabasePath  string
	EncryptionKey []byte
	AllowOrigins  string
	Environment   string
}

var C *Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	keyHex := os.Getenv("PROXERA_ENCRYPTION_KEY")
	if keyHex == "" {
		log.Fatal("PROXERA_ENCRYPTION_KEY is required (64-char hex = 32-byte AES-256 key)")
	}
	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil || len(keyBytes) != 32 {
		log.Fatalf("PROXERA_ENCRYPTION_KEY must be a 64-char hex string (got %d chars): %v", len(keyHex), err)
	}

	port := getEnv("PORT", "8080")
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("PORT must be a valid integer: %v", err)
	}

	C = &Config{
		Port:          port,
		DatabasePath:  getEnv("DATABASE_PATH", "./data/proxera.db"),
		EncryptionKey: keyBytes,
		AllowOrigins:  getEnv("ALLOW_ORIGINS", "http://localhost:5173"),
		Environment:   getEnv("ENVIRONMENT", "development"),
	}

	fmt.Printf("Proxera backend starting on :%s (env=%s)\n", C.Port, C.Environment)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
