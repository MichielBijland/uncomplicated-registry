package cmd

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"github.com/MichielBijland/uncomplicated-registry/internal/module"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagModuleNameSpace          string
	flagModuleName               string
	flagModuleProvider           string
	flagModuleVersion            string
	flagVersionConstraintsRegex  string
	flagVersionConstraintsSemver string
)

var (
	versionConstraintsRegex  *regexp.Regexp
	versionConstraintsSemver version.Constraints
)

func init() {
	rootCmd.AddCommand(uploadCmd)
	uploadCmd.Flags().StringVar(&flagModuleNameSpace, "namespace", "", "The namespace of the module")
	uploadCmd.MarkFlagRequired("namespace")
	uploadCmd.Flags().StringVar(&flagModuleName, "name", "", "The name of the module")
	uploadCmd.MarkFlagRequired("name")
	uploadCmd.Flags().StringVar(&flagModuleProvider, "provider", "", "The provider of the module")
	uploadCmd.MarkFlagRequired("provider")
	uploadCmd.Flags().StringVar(&flagModuleVersion, "version", "", "The version of the module")
	uploadCmd.MarkFlagRequired("version")
	uploadCmd.Flags().StringVar(&flagVersionConstraintsRegex, "version-constraints-regex", "", `Limit the module versions that are eligible for upload with a regex that a version has to match.
Can be combined with the -version-constraints-semver flag`)
	uploadCmd.Flags().StringVar(&flagVersionConstraintsSemver, "version-constraints-semver", "", `Limit the module versions that are eligible for upload with version constraints.
The version string has to be formatted as a string literal containing one or more conditions, which are separated by commas.
Can be combined with the -version-constrained-regex flag`)
}

var uploadCmd = &cobra.Command{
	Short:        "Upload a module to the registry",
	Use:          "upload [flags]] MODULE",
	SilenceUsage: true,
	RunE:         uploadModule,
}

func uploadModule(cmd *cobra.Command, args []string) error {
	storageBackend, err := setupStorage(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to setup storage")
	}

	if len(args) == 0 {
		return fmt.Errorf("missing argument")
	}

	if _, err := os.Stat(args[0]); errors.Is(err, os.ErrNotExist) {
		return err
	}

	// constuct metadata and validate
	metadata := module.Metadata{
		Namespace: flagModuleNameSpace,
		Name:      flagModuleName,
		Provider:  flagModuleProvider,
		Version:   flagModuleVersion,
	}
	err = metadata.Validate()
	if err != nil {
		return err
	}

	// Validate the semver version constraints
	if flagVersionConstraintsSemver != "" {
		constraints, err := version.NewConstraint(flagVersionConstraintsSemver)
		if err != nil {
			return err
		}
		versionConstraintsSemver = constraints
	}

	// Validate the regex version constraints
	if flagVersionConstraintsRegex != "" {
		constraints, err := regexp.Compile(flagVersionConstraintsRegex)
		if err != nil {
			return fmt.Errorf("invalid regex given: %v", err)
		}
		versionConstraintsRegex = constraints
	}

	return archiveModules(args[0], metadata, storageBackend)
}
