package configs

import (
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func LoggerConfig() logger.Config {
	return logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}
}
