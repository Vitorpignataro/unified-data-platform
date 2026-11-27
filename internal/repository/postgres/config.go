package postgres

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		Host:     os.Getenv("WAREHOUSE_HOST"),
		Database: os.Getenv("POSTGRES_DB"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
	}

	if cfg.Host == "" {
		cfg.Host = "127.0.0.1"
	}

	port := os.Getenv("WAREHOUSE_PORT")
	if port == "" {
		port = "5433"
	}

	parsedPort, err := strconv.Atoi(port)
	if err != nil || parsedPort < 1 || parsedPort > 65535 {
		return Config{}, fmt.Errorf(
			"WAREHOUSE_PORT inválida: %q",
			port,
		)
	}

	cfg.Port = parsedPort

	if cfg.Database == "" ||
		cfg.User == "" ||
		cfg.Password == "" {
		return Config{}, fmt.Errorf(
			"credenciais PostgreSQL incompletas",
		)
	}

	return cfg, nil
}
