package cmd

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/MichielBijland/uncomplicated-registry/internal/module"
	"github.com/MichielBijland/uncomplicated-registry/internal/utils"

	"github.com/hashicorp/go-version"
)

func archiveModules(root string, metadata module.Metadata, storage module.Storage) error {
	return processModule(root, metadata, storage)
}

func processModule(path string, metadata module.Metadata, storage module.Storage) error {

	// Check if the module meets version constraints
	if versionConstraintsSemver != nil {
		ok, err := meetsSemverConstraints(&metadata)
		if err != nil {
			return err
		} else if !ok {
			// Skip the module, as it didn't pass the version constraints
			logger.Info().Str("metadata", metadata.String()).Msg("module doesn't meet semver version constraints, skipped")
			return nil
		}
	}

	if versionConstraintsRegex != nil {
		if !meetsRegexConstraints(&metadata) {
			// Skip the module, as it didn't pass the regex version constraints
			logger.Info().Str("metadata", metadata.String()).Msg("module doesn't meet regex version constraints, skipped")
			return nil
		}
	}

	ctx := context.Background()
	if res, err := storage.GetModule(ctx, metadata.Namespace, metadata.Name, metadata.Provider, metadata.Version); err == nil {
		logger.Error().Str("download_url", res.DownloadURL).Msg("module already exists")
		return errors.New("module already exists")
	}

	moduleRoot := filepath.Dir(path)

	buf, err := utils.ArchiveModule(moduleRoot, logger)
	if err != nil {
		return err
	}

	res, err := storage.UploadModule(ctx, metadata.Namespace, metadata.Name, metadata.Provider, metadata.Version, buf)
	if err != nil {
		return err
	}

	logger.Info().Str("download_url", res.DownloadURL).Msg("module successfully uploaded")

	return nil

}

// meetsSemverConstraints checks whether a module version matches the semver version constraints.
// Returns an unrecoverable error if there's an internal error. Otherwise it returns a boolean indicating if the module meets the constraints
func meetsSemverConstraints(meta *module.Metadata) (bool, error) {
	v, err := version.NewSemver(meta.Version)
	if err != nil {
		return false, err
	}

	return versionConstraintsSemver.Check(v), nil
}

// meetsRegexConstraints checks whether a module version matches the regex.
// Returns a boolean indicating if the module meets the constraints
func meetsRegexConstraints(meta *module.Metadata) bool {
	return versionConstraintsRegex.MatchString(meta.Version)
}
