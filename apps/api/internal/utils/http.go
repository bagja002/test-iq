package utils

import "github.com/gofiber/fiber/v3"

type APIError struct {
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func RespondError(c fiber.Ctx, status int, message string, details string) error {
	return c.Status(status).JSON(APIError{
		Message: message,
		Details: details,
	})
}

func RespondMessage(c fiber.Ctx, status int, message string, extra fiber.Map) error {
	payload := fiber.Map{
		"message": message,
	}

	for key, value := range extra {
		payload[key] = value
	}

	return c.Status(status).JSON(payload)
}
