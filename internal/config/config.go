package config

import (
	"os"
)

type Config struct {
	Port       string
	BaseURL    string
	DBPath     string
	HashSalt   string
	RL_Enabled bool
	RL_RPS     int
	RL_BURST   int
}

func defaultIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}

func LoadFromEnv() *Config {
	// minimal parsing; ints handled with defaults
	port := defaultIfEmpty(os.Getenv("PORT"), "8080")
	base := defaultIfEmpty(os.Getenv("BASE_URL"), "http://localhost:8080")
	db := defaultIfEmpty(os.Getenv("DB_PATH"), "./data/qr.db")
	salt := defaultIfEmpty(os.Getenv("HASH_SALT"), "change-me")

	rlEnabled := true
	// defaults
	rlRPS := 5
	rlBurst := 10

	return &Config{
		Port:       port,
		BaseURL:    base,
		DBPath:     db,
		HashSalt:   salt,
		RL_Enabled: rlEnabled,
		RL_RPS:     rlRPS,
		RL_BURST:   rlBurst,
	}
}
