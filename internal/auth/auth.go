package auth

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/rs/zerolog"
)

func Middleware(logger zerolog.Logger, providers ...Provider) func(c *fiber.Ctx) error {
	// NO-OP, as there are no providers defined
	if len(providers) == 0 {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return keyauth.New(keyauth.Config{
		Validator: func(c *fiber.Ctx, key string) (bool, error) {

			for _, provider := range providers {
				success, err := provider.Verify(c, key)
				if err != nil {
					logger.Error().Str("provider", provider.String()).Err(err).Msg("failed to verify token")
				} else {
					return success, nil
				}
			}

			return false, keyauth.ErrMissingOrMalformedAPIKey
		},
	})
}
