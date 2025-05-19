package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var GoogleOAuthConfig *oauth2.Config

func init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, loading from system environment variables")
	} else {
		log.Println(".env file loaded successfully")
	}

	// Log environment variables for debugging
	log.Println("GOOGLE_CLIENT_ID:", os.Getenv("GOOGLE_CLIENT_ID"))
	log.Println("GOOGLE_CLIENT_SECRET:", os.Getenv("GOOGLE_CLIENT_SECRET"))
	log.Println("GOOGLE_REDIRECT_URI:", os.Getenv("GOOGLE_REDIRECT_URI"))
	log.Println("GOOGLE_APPLICATION_CREDENTIALS:", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

	// Initialize GoogleOAuthConfig after loading environment variables
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URI"),
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

type Config struct {
	Port      string
	DB        *pgxpool.Pool
	JWTSecret string
}

func InitializeDB() (*gorm.DB, error) {
	// Replace with your actual database configuration
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// LoadConfig loads environment variables and connects to PostgreSQL
func LoadConfig() (*Config, error) {
	// Load from .env file if available
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, loading from system environment variables")
	}

	//port := os.Getenv("PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	// PostgreSQL config
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	log.Println("DB_USER:", dbUser)
	log.Println("DB_PASSWORD:", dbPassword)
	log.Println("DB_NAME:", dbName)
	// Form DB URL
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// Connect with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	// Return config
	return &Config{
		Port:      os.Getenv("PORT"),
		DB:        db,
		JWTSecret: jwtSecret,
	}, nil
}
