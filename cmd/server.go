package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/MichielBijland/uncomplicated-registry/internal/auth"
	"github.com/MichielBijland/uncomplicated-registry/internal/storage"
	"github.com/rs/zerolog"

	"github.com/pkg/errors"

	"golang.org/x/sync/errgroup"

	"github.com/MichielBijland/uncomplicated-registry/internal/discovery"
	"github.com/MichielBijland/uncomplicated-registry/internal/module"

	"github.com/spf13/cobra"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	apiVersion = "v1"
)

var (
	prefix        = fmt.Sprintf("/%s", apiVersion)
	prefixModules = fmt.Sprintf("%s/modules", prefix)
)

var (
	// General server options.
	flagTLSCertFile string
	flagTLSKeyFile  string
	flagListenAddr  string

	// Login options.
	flagLoginClient     string
	flagLoginScopes     []string
	flagLoginGrantTypes []string
	flagLoginAuthz      string
	flagLoginToken      string
	flagLoginPorts      []int

	// Static auth.
	flagAuthStaticTokens []string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the server component",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		group, ctx := errgroup.WithContext(ctx)

		server, err := createFiber(logger)
		if err != nil {
			return errors.Wrap(err, "failed to setup server")
		}

		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// Signal handler.
		group.Go(func() error {
			select {
			case <-sigint:
				cancel()
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})

		// Server handler.
		group.Go(func() error {
			<-ctx.Done()

			if err := server.ShutdownWithContext(ctx); err != nil {
				if err != context.Canceled {
					logger.Error().Err(err).Msg("failed to terminate server")
				}
			}

			return nil
		})

		// Main server.
		group.Go(func() error {
			sublogger := logger.With().Str("listen", flagListenAddr).Logger()
			sublogger.Info().Msg("starting server")
			defer sublogger.Info().Msg("shutting down server")

			if flagTLSCertFile != "" || flagTLSKeyFile != "" {
				if err := server.ListenTLS(flagListenAddr, flagTLSCertFile, flagTLSKeyFile); err != nil {
					if err != http.ErrServerClosed {
						return err
					}
				}
			} else {
				if err := server.Listen(flagListenAddr); err != nil {
					if err != http.ErrServerClosed {
						return err
					}
				}
			}
			return nil
		})

		return group.Wait()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// General options.
	serverCmd.Flags().StringVar(&flagTLSKeyFile, "tls-key-file", "", "TLS private key to serve")
	serverCmd.Flags().StringVar(&flagTLSCertFile, "tls-cert-file", "", "TLS certificate to serve")
	serverCmd.Flags().StringVar(&flagListenAddr, "listen-address", ":5601", "Address to listen on")
	// Static auth options.
	serverCmd.Flags().StringSliceVar(&flagAuthStaticTokens, "auth-static-token", nil, "Static API token to protect the uncomplicated-registry")

	// Terraform Login Protocol options.
	serverCmd.Flags().StringVar(&flagLoginClient, "login-client", "", "The client_id value to use when making requests")
	serverCmd.Flags().StringSliceVar(&flagLoginGrantTypes, "login-grant-types", []string{"authz_code"}, "An array describing a set of OAuth 2.0 grant types")
	serverCmd.Flags().StringVar(&flagLoginAuthz, "login-authz", "", "The server's authorization endpoint")
	serverCmd.Flags().StringVar(&flagLoginToken, "login-token", "", "The server's token endpoint")
	serverCmd.Flags().IntSliceVar(&flagLoginPorts, "login-ports", []int{10000, 10010}, "Inclusive range of TCP ports that Terraform may use")
	serverCmd.Flags().StringSliceVar(&flagLoginScopes, "login-scopes", nil, "List of scopes")
}

func setupStorage(ctx context.Context) (storage.Storage, error) {
	switch {
	case flagS3Bucket != "":
		return storage.NewS3Storage(ctx,
			flagS3Bucket,
			storage.WithS3StorageBucketPrefix(flagS3Prefix),
			storage.WithS3StorageBucketRegion(flagS3Region),
			storage.WithS3StorageBucketEndpoint(flagS3Endpoint),
			storage.WithS3StoragePathStyle(flagS3PathStyle),
			storage.WithS3ArchiveFormat(storage.DefaultModuleArchiveFormat),
			storage.WithS3StorageSignedUrlExpiry(flagS3SignedURLExpiry),
		)
	default:
		return nil, errors.New("please specify a valid storage provider")
	}
}

func createFiber(logger zerolog.Logger) (*fiber.App, error) {
	app := fiber.New()

	app.Use(recover.New())
	app.Use(etag.New())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger,
	}))

	if err := registerDiscovery(app); err != nil {
		return nil, err
	}

	s, err := setupStorage(context.TODO())
	if err != nil {
		return nil, err
	}

	if err := registerModule(app, s); err != nil {
		return nil, err
	}

	return app, nil
}

func registerDiscovery(app *fiber.App) error {

	options := []discovery.Option{
		discovery.WithModulesV1(fmt.Sprintf("%s/", prefixModules)),
	}

	if flagLoginClient != "" {
		login := &discovery.LoginV1{
			Client: flagLoginClient,
		}

		if flagLoginGrantTypes != nil {
			login.GrantTypes = flagLoginGrantTypes
		}

		if flagLoginAuthz != "" {
			login.Authz = flagLoginAuthz
		}

		if flagLoginToken != "" {
			login.Token = flagLoginToken
		}

		if flagLoginPorts != nil {
			login.Ports = flagLoginPorts
		}

		if flagLoginScopes != nil {
			login.Scopes = flagLoginScopes
		}

		options = append(options, discovery.WithLoginV1(login))
	}

	discovery := discovery.New(options...)

	app.Head("/.well-known/terraform.json", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, no-cache, must-revalidate")
		return c.JSON(discovery)
	})

	app.Get("/.well-known/terraform.json", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderCacheControl, "public, no-cache, must-revalidate")
		return c.JSON(discovery)
	})

	return nil
}

func registerModule(app *fiber.App, s storage.Storage) error {
	service := module.NewService(s)

	api := app.Group(prefixModules)
	api.Use(authMiddleware(logger))

	module.Register(service, api)

	return nil
}

func authMiddleware(logger zerolog.Logger) func(c *fiber.Ctx) error {
	var providers []auth.Provider

	if flagAuthStaticTokens != nil {
		providers = append(providers, auth.NewStaticProvider(logger, flagAuthStaticTokens...))
	}

	return auth.Middleware(logger, providers...)
}
