package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	DBSSLMode    string
	JWTSecret    string
	ServerPort   string
	FrontendURL  string
	CookieSecure bool
}

func Load() *Config {
	godotenv.Load()

	isProduction := os.Getenv("APP_ENV") == "production"

	jwtSecret := getEnv("JWT_SECRET", "")
	dbPassword := getEnv("DB_PASSWORD", "")

	if isProduction {
		if jwtSecret == "" {
			log.Fatal("JWT_SECRET wajib diset di production")
		}
		if dbPassword == "" {
			log.Fatal("DB_PASSWORD wajib diset di production")
		}
	}

	if jwtSecret == "" {
		jwtSecret = "dev-secret"
	}
	if dbPassword == "" {
		dbPassword = "wtrlab_secret"
	}

	return &Config{
		DBHost:       getEnv("DB_HOST", "localhost"),
		DBPort:       getEnv("DB_PORT", "5432"),
		DBUser:       getEnv("DB_USER", "wtrlab"),
		DBPassword:   dbPassword,
		DBName:       getEnv("DB_NAME", "wtrlab"),
		DBSSLMode:    getEnv("DB_SSLMODE", "disable"),
		JWTSecret:    jwtSecret,
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
		CookieSecure: getEnv("COOKIE_SECURE", "true") == "true",
	}
}

func (c *Config) DSN() string {
	return "host=" + c.DBHost +
		" port=" + c.DBPort +
		" user=" + c.DBUser +
		" password=" + c.DBPassword +
		" dbname=" + c.DBName +
		" sslmode=" + c.DBSSLMode
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
