package configs

import "github.com/gofiber/fiber/v2/middleware/cors"

func CorsConfig() cors.Config {
    return cors.Config{
        AllowOrigins:     "http://localhost:3000, https://webbuilderv2.vercel.app, https://basilisk-needed-usually.ngrok-free.app",
        AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
        AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
        AllowCredentials: true,
    }
}