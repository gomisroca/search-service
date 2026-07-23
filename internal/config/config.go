package config

import (
	"os"
	"strconv"
)
 
type Config struct {
	Port string
	APIKey string

	// How many documents the index will hold, 0 = no limit
	MaxDocuments int

	// Used when no limit is provided
	DefaultSearchLimit int

	MaxSearchLimit int

	MinTokenLength int
}

func Load() Config {
	return Config{
		Port:               getEnv("PORT", "8080"),
		APIKey:             getEnv("API_KEY", ""),
		MaxDocuments:       getEnvInt("MAX_DOCUMENTS", 0),
		DefaultSearchLimit: getEnvInt("DEFAULT_SEARCH_LIMIT", 10),
		MaxSearchLimit:     getEnvInt("MAX_SEARCH_LIMIT", 100),
		MinTokenLength:     getEnvInt("MIN_TOKEN_LENGTH", 2),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
 
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
