// Package bootstrap holds the process start-up steps shared by the service
// binaries (cmd/api, cmd/grpc).
package bootstrap

import (
	"log"
	"os"

	"my-go-app/internal/database"
	"my-go-app/internal/repositories"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
)

// LoadEnv reads .env (if present) and configures the Clerk SDK from
// CLERK_SECRET_KEY. A missing .env is not an error: real deployments set the
// environment directly.
func LoadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: Could not load .env file, using system environment variables")
	}
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
}

// InitRepositories loads the environment, opens the database connection and
// builds the repository set. It exits the process if the database is
// unreachable.
func InitRepositories() *repositories.RepositoriesInterface {
	LoadEnv()

	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Failed to initialize database connection pool:", err)
	}
	return database.NewRepositories(db)
}
