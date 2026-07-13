package config

import (
	"crypto/rand"
	"encoding/hex"
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

func generateSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("failed to generate random secret: %v", err)
	}
	return hex.EncodeToString(b)
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables as-is")
	}

	isProduction := os.Getenv("APP_ENV") == "production"

	jwtSecret := os.Getenv("JWT_SECRET")
	dbPassword := os.Getenv("DB_PASSWORD")

	if isProduction {
		if jwtSecret == "" {
			log.Fatal("JWT_SECRET wajib diset di production")
		}
		if dbPassword == "" {
			log.Fatal("DB_PASSWORD wajib diset di production")
		}
	}

	if jwtSecret == "" {
		jwtSecret = generateSecret()
		log.Println("[WARN] JWT_SECRET tidak diset — menggunakan secret random (sementara)." +
			" Set JWT_SECRET di .env untuk persistensi antar restart.")
	}
	if dbPassword == "" {
		dbPassword = "wtrlab_secret"
		log.Println("[WARN] DB_PASSWORD tidak diset — menggunakan default 'wtrlab_secret'." +
			" Set DB_PASSWORD di .env untuk keamanan.")
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
