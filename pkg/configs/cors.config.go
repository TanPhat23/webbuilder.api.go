package configs

import (
	"errors"
	"os"
	"strings"

	"github.com/gofiber/fiber/v3/middleware/cors"
)

const defaultDevAllowedOrigins = "http://localhost:3000, https://webbuilderv2.vercel.app"

// AllowedOrigins returns the configured CORS origin allowlist and refuses an
// empty ALLOWED_ORIGINS in production environments.
func AllowedOrigins() ([]string, error) {
	allowOrigins := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if allowOrigins == "" {
		if env == "production" || env == "prod" {
			return nil, errors.New("ALLOWED_ORIGINS must be set when ENV is production")
		}
		allowOrigins = defaultDevAllowedOrigins
	}

	var allowOriginsList []string
	for origin := range strings.SplitSeq(allowOrigins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			allowOriginsList = append(allowOriginsList, trimmed)
		}
	}
	if len(allowOriginsList) == 0 {
		return nil, errors.New("ALLOWED_ORIGINS must contain at least one origin")
	}
	return allowOriginsList, nil
}

func CorsConfig() cors.Config {
	allowOriginsList, err := AllowedOrigins()
	if err != nil {
		allowOriginsList = nil
	}
	return cors.Config{
		AllowOrigins:     allowOriginsList,
		AllowMethods:     []string{"GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}
}
