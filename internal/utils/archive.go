package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rs/zerolog"
)

// TODO: make this configurable and use this as default
const allowedFilesRegex = `^.*\.(tf\.json|tftpl|tf)|README(\.md){0,1}|LICENSE$`

func ArchiveModule(root string, logger zerolog.Logger) (io.Reader, error) {
	allowed, err := regexp.Compile(allowedFilesRegex)
	if err != nil {
		logger.Fatal().Err(err).Msg("invalid regex given")
		return nil, err
	}

	buf := new(bytes.Buffer)
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(root); err != nil {
		return buf, fmt.Errorf("unable to tar files - %v", err.Error())
	}

	gw := gzip.NewWriter(buf)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files
		if !fi.Mode().IsRegular() {
			return nil
		}

		// only allow Nested modules at root level
		if fi.IsDir() {
			if path != root && fi.Name() == "modules" {
				logger.Info().Msg("Skipping to deeply nested child module: " + fi.Name())
				return filepath.SkipDir
			}
		}

		// only allow files that match the allowedFilesRegex pattern
		if !fi.IsDir() {
			if !allowed.MatchString(fi.Name()) {
				logger.Info().Msg("Skipping file: " + fi.Name() + " as it does not match the allowedFilesRegex pattern")
				return nil
			}
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring
		header.Name = strings.TrimPrefix(strings.Replace(path, root, "", -1), string(filepath.Separator))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		data, err := os.Open(path)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, data); err != nil {
			return err
		}

		// manually close here after each file operation; deferring would cause each file close
		// to wait until all operations have completed.
		data.Close()

		return nil
	})

	return buf, err
}
