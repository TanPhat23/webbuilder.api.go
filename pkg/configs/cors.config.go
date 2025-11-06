package configs

import (
	"os"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

func CorsConfig() cors.Config {
	allowOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:3000, https://webbuilderv2.vercel.app, https://basilisk-needed-usually.ngrok-free.app"
	}
    return cors.Config{
        AllowOrigins:     allowOrigins,
        AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
        AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
        AllowCredentials: true,
    }
}
