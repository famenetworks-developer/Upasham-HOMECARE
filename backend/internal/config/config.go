package config

import (
	"os"
)

type Config struct {
	Port         string
	DBURL        string
	JWTSecret    string
	JitsiDomain  string
	Env          string
	AllowedCORS  string
	GeneralPrice float64 // default BDT 20
}

func LoadConfig() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		// Default to local PostgreSQL on 5433 with upasham_db
		dbURL = "postgres://postgres:postgres@127.0.0.1:5433/upasham_db?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "upasham-connected-care-dhaka-secret-key-2026-bmdc"
	}

	jitsiDomain := os.Getenv("JITSI_DOMAIN")
	if jitsiDomain == "" {
		jitsiDomain = "meet.jit.si"
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	return &Config{
		Port:         port,
		DBURL:        dbURL,
		JWTSecret:    jwtSecret,
		JitsiDomain:  jitsiDomain,
		Env:          env,
		AllowedCORS:  "*",
		GeneralPrice: 20.0,
	}
}
