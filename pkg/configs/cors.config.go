package configs

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3/middleware/cors"
)

func CorsConfig() cors.Config {
	allowOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:3000, https://webbuilderv2.vercel.app, https://basilisk-needed-usually.ngrok-free.app"
	}

	allowOriginsList := []string{}
	for origin := range strings.SplitSeq(allowOrigins, ",") {
		allowOriginsList = append(allowOriginsList, strings.TrimSpace(origin))
	}
	return cors.Config{
		AllowOrigins:     allowOriginsList,
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
}
