package configs

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3/middleware/logger"
)

// SetupLogging installs a JSON slog handler as the process default so that
// hosts such as Railway can parse the level and attributes of each line.
func SetupLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func LoggerConfig() logger.Config {
	return logger.Config{
		Format: "${status} - ${method} ${path} user:${locals:userId} req:${respHeader:X-Request-ID}\n",
	}
}
