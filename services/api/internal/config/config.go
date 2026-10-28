package config

import (
	"os"
)

type Config struct {
	Port                string
	DatabaseURL         string
	ValkeyURL           string
	NatsURL             string
	JWTSecret           string
	R2Endpoint          string
	R2AccessKey         string
	R2SecretKey         string
	R2Bucket            string
	CloudflareStreamKey string
}

func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/gvp?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "super-secret-jwt-key-for-development-32chars!"
	}

	return &Config{
		Port:        port,
		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		R2Bucket:    "gvp-media-vod",
	}
}
