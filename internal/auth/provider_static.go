package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
	"github.com/rs/zerolog"
)

type StaticProvider struct {
	tokens [][32]byte
}

func (p *StaticProvider) String() string { return "static" }

func (p *StaticProvider) Verify(ctx *fiber.Ctx, key string) (bool, error) {
	hashedKey := sha256.Sum256([]byte(key))
	for _, validToken := range p.tokens {
		if subtle.ConstantTimeCompare(validToken[:], hashedKey[:]) == 1 {
			return true, nil
		}
	}

	return false, keyauth.ErrMissingOrMalformedAPIKey
}

func NewStaticProvider(logger zerolog.Logger, tokens ...string) Provider {
	// spf13/viper and spf13/pflag currently do not support reading multiple values from environment variables and
	// extracting them into a StringSlice/StringArray.
	// This workaround extracts comma-separated tokens into separate tokens
	//
	// See https://github.com/spf13/viper/issues/339 and https://github.com/spf13/viper/issues/380
	var parsed [][32]byte
	for _, t := range tokens {
		if strings.ContainsAny(t, ",") {
			split := strings.Split(t, ",")
			for _, s := range split {
				if s == "" {
					// Skip empty strings occurring due to splitting csv values like "test,"
					continue
				}
				parsed = append(parsed, sha256.Sum256([]byte(s)))
			}
		} else {
			parsed = append(parsed, sha256.Sum256([]byte(t)))
		}
	}

	return &StaticProvider{
		tokens: parsed,
	}
}
