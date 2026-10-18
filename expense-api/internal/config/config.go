package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port      string
	DBPath    string
	LogPath   string
	MaxAmount float64
}

// Load reads environment variables with fallback defaults.
func Load() *Config {
	return &Config{
		Port:      getEnv("PORT", "8080"),
		DBPath:    getEnv("DB_PATH", "./data/expenses.db"),
		LogPath:   getEnv("LOG_PATH", "./data/app.log"),
		MaxAmount: getEnvFloat("MAX_EXPENSE_AMOUNT", 100000000.00),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil && f > 0 {
			return f
		}
	}
	return defaultVal
}
