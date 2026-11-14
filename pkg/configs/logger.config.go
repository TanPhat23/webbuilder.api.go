package configs

import (
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func LoggerConfig() logger.Config {
	return logger.Config{
		Format: "${status} - ${method} ${path} user:${locals:userId}\n",
	}
}
