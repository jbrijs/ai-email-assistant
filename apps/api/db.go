package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var dbPool *pgxpool.Pool

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

// InitDB initializes the database connection pool
func InitDB() error {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using system environment variables")
	}

	config := DatabaseConfig{
		Host:     getenvDB("DB_HOST", "localhost"),
		Port:     getenvDB("DB_PORT", "5432"),
		User:     getenvDB("DB_USER", "postgres"),
		Password: getenvDB("DB_PASSWORD", "postgres"),
		Database: getenvDB("DB_NAME", "inboxai"),
		SSLMode:  getenvDB("DB_SSLMODE", "disable"),
	}

	// Build connection string
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.User, config.Password, config.Host, config.Port, config.Database, config.SSLMode,
	)

	// Create connection pool configuration
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	dbPool = pool
	log.Printf("Database connected successfully to %s:%s/%s", config.Host, config.Port, config.Database)

	return nil
}



// GetDB returns the database connection pool
func GetDB() *pgxpool.Pool {
	return dbPool
}

// CloseDB closes the database connection pool
func CloseDB() {
	if dbPool != nil {
		dbPool.Close()
		log.Println("Database connection pool closed")
	}
}

// Helper function to get environment variable with default
func getenvDB(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
