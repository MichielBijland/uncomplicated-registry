package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/rs/zerolog"
)

const (
	projectName = "uncomplicated-registry"
	envPrefix   = "UNCOMPLICATED_REGISTRY"
)

var (
	flagDebug bool

	// S3 options.
	flagS3Bucket          string
	flagS3Prefix          string
	flagS3Region          string
	flagS3Endpoint        string
	flagS3PathStyle       bool
	flagS3SignedURLExpiry time.Duration
)

var (
	logger zerolog.Logger
)

var rootCmd = &cobra.Command{
	Use:           projectName,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := initializeConfig(cmd); err != nil {
			return err
		}

		logger = setupLogger()

		if flagDebug {
			logger.Info().Msg("debug mode enabled")
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&flagS3Bucket, "storage-s3-bucket", "", "S3 bucket to use for the registry")
	rootCmd.PersistentFlags().StringVar(&flagS3Prefix, "storage-s3-prefix", "", "S3 bucket prefix to use for the registry")
	rootCmd.PersistentFlags().StringVar(&flagS3Region, "storage-s3-region", "", "S3 bucket region to use for the registry")
	rootCmd.PersistentFlags().StringVar(&flagS3Endpoint, "storage-s3-endpoint", "", "S3 bucket endpoint URL (required for MINIO)")
	rootCmd.PersistentFlags().BoolVar(&flagS3PathStyle, "storage-s3-pathstyle", false, "S3 use PathStyle (required for MINIO)")
	rootCmd.PersistentFlags().DurationVar(&flagS3SignedURLExpiry, "storage-s3-signedurl-expiry", 30*time.Second, "Generate S3 signed URL valid for X seconds. Only meaningful if used in combination with --storage-s3-signedurl")
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()
	v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()
	bindFlags(cmd, v)
	return nil
}

func setupLogger() zerolog.Logger {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if flagDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	return logger
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		envVarSuffix := strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_"))
		if err := v.BindEnv(f.Name, fmt.Sprintf("%s_%s", envPrefix, envVarSuffix)); err != nil {
			panic(fmt.Errorf("failed to bind key to environment variable: %w", err))
		}
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				panic(fmt.Errorf("failed to set value of flag: %w", err))
			}
		}
	})
}
