package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	HTTPAddress string
	StaticDir   string
}

func Load() (Config, error) {
	cfg := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		HTTPAddress: os.Getenv("HTTP_ADDRESS"),
		StaticDir:   os.Getenv("STATIC_DIR"),
	}
	if cfg.HTTPAddress == "" {
		cfg.HTTPAddress = ":8080"
	}
	if cfg.StaticDir == "" {
		cfg.StaticDir = "./frontend"
	}

	if cfg.DatabaseURL == "" {
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		sslmode := os.Getenv("DB_SSLMODE")

		if user == "" {
			return Config{}, fmt.Errorf("DB_USER is required")
		}
		if password == "" {
			return Config{}, fmt.Errorf("DB_PASSWORD is required")
		}
		if name == "" {
			return Config{}, fmt.Errorf("DB_NAME is required")
		}
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "5432"
		}
		if sslmode == "" {
			sslmode = "disable"
		}

		dbURL := url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(user, password),
			Host:   net.JoinHostPort(host, port),
			Path:   "/" + name,
			RawQuery: url.Values{
				"sslmode": {sslmode},
			}.Encode(),
		}
		cfg.DatabaseURL = dbURL.String()
	}

	return cfg, nil
}
