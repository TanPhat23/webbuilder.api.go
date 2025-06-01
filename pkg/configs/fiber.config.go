package configs

import "github.com/gofiber/fiber/v2"

func FiberConfig() fiber.Config {
	return fiber.Config{
		Prefork:       false,
		CaseSensitive: true,
		StrictRouting: true,
		ServerHeader:  "Fiber",
		AppName:       "Webbuilder v1.0.1",
	}
}