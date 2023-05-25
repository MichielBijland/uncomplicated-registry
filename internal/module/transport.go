package module

import (
	"github.com/gofiber/fiber/v2"
)

func Register(svc Service, router fiber.Router) {
	router.Get("/:namespace/:name/:provider/versions", listEndpoint(svc))
	router.Get("/:namespace/:name/:provider/:version/download", downloadEndpoint(svc))
}

func errorHandler(c *fiber.Ctx, err error) error {
	errors := []string{err.Error()}
	response := fiber.Map{
		"errors": errors,
	}
	switch err.(type) {
	case *fiber.Error:
		return c.Status(err.(*fiber.Error).Code).JSON(response)
	default:
		return c.Status(fiber.StatusInternalServerError).JSON(response)
	}
}

func notFoundHandler(c *fiber.Ctx) error {
	errors := []string{"Not Found"}
	response := fiber.Map{
		"errors": errors,
	}
	return c.Status(fiber.StatusNotFound).JSON(response)
}
