package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

// InitDB initializes the database connection pool once at startup
func InitDB() error {
	var err error
	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}

	// Configure connection pool for optimal performance
	DB.SetMaxOpenConns(100)                 // Maximum number of open connections
	DB.SetMaxIdleConns(25)                 // Maximum number of idle connections  

	// Test the connection
	if err = DB.Ping(); err != nil {
		return err
	}

	log.Println("Database connection pool initialized successfully")
	return nil
}

// GetDB returns the shared database connection
func GetDB() *sql.DB {
	return DB
}
