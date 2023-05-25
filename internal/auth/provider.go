package auth

import (
	"github.com/gofiber/fiber/v2"
)

type Provider interface {
	Verify(ctx *fiber.Ctx, token string) (bool, error)
	String() string
}
