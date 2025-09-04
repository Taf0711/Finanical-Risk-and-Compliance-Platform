package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns a CORS middleware with default configuration
func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000, http://localhost:8080, http://127.0.0.1:3000, http://127.0.0.1:8080",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
	})
}

// CORSWithConfig returns a CORS middleware with custom configuration
func CORSWithConfig(config cors.Config) fiber.Handler {
	return cors.New(config)
}
