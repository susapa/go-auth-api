package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                   string
	DatabaseURL            string
	JWTSecret              string
	JWTExpiryHours         int
	RefreshTokenExpiryDays int
	AllowedOrigins         string
	AzureStorageConnStr    string
	AzureStorageContainer  string
	AzureStorageAccount    string
}

var C Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading from environment")
	}

	jwtExpiry, _ := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	refreshExpiry, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRY_DAYS", "7"))

	C = Config{
		Port:                   getEnv("PORT", "8080"),
		DatabaseURL:            mustGetEnv("DATABASE_URL"),
		JWTSecret:              mustGetEnv("JWT_SECRET"),
		JWTExpiryHours:         jwtExpiry,
		RefreshTokenExpiryDays: refreshExpiry,
		AllowedOrigins:         getEnv("ALLOWED_ORIGINS", "http://localhost:4200"),
		AzureStorageConnStr:    mustGetEnv("AZURE_STORAGE_CONNECTION_STRING"),
		AzureStorageContainer:  getEnv("AZURE_STORAGE_CONTAINER", "slips"),
		AzureStorageAccount:    mustGetEnv("AZURE_STORAGE_ACCOUNT_NAME"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s is not set", key)
	}
	return v
}
