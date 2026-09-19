package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	MongoURI      string
	MongoDBName   string
	RedisURL      string
	RedisPassword string
	JWTSecret     string
	ClientOrigin  string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found or error loading, falling back to environment variables")
	}

	port := getEnv("PORT", "8080")
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	mongoDBName := getEnv("MONGO_DB_NAME", "livepolling")
	redisURL := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	jwtSecret := getEnv("JWT_SECRET", "super-secure-guvi-internship-jwt-secret-key-2026")
	clientOrigin := getEnv("CLIENT_ORIGIN", "http://localhost:5173")

	return &Config{
		Port:          port,
		MongoURI:      mongoURI,
		MongoDBName:   mongoDBName,
		RedisURL:      redisURL,
		RedisPassword: redisPassword,
		JWTSecret:     jwtSecret,
		ClientOrigin:  clientOrigin,
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
