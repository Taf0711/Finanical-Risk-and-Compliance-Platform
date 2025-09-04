package middleware

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Logger returns a logger middleware with default configuration
func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     os.Stdout,
	})
}

// LoggerWithConfig returns a logger middleware with custom configuration
func LoggerWithConfig(config logger.Config) fiber.Handler {
	return logger.New(config)
}

// FileLogger returns a logger middleware that writes to a file
func FileLogger(filename string) fiber.Handler {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency} - ${error}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
		Output:     file,
	})
}

// RequestIDLogger returns a logger middleware with request ID
func RequestIDLogger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${locals:requestid} ${status} - ${method} ${path} - ${ip} - ${latency}\n",
		TimeFormat: time.RFC3339,
		TimeZone:   "UTC",
	})
}
