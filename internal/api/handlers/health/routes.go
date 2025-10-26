package health

import "github.com/gofiber/fiber/v2"

func SetupRoutes(router fiber.Router, handler *HealthHandler) {
	router.Get("", handler.Get)
	router.Get("/", handler.Get)
}