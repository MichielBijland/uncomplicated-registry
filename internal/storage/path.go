package storage

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/MichielBijland/uncomplicated-registry/internal/core"
)

const (
	internalModuleType = moduleType("modules")
)

type moduleType string

// modulePathPrefix returns a <prefix>/modules/<namespace>/<name>/<provider> prefix
func modulePathPrefix(prefix, namespace, name, provider string) string {
	return path.Join(prefix, string(internalModuleType), namespace, name, provider)
}

func modulePath(prefix, namespace, name, provider, version, archiveFormat string) string {
	f := fmt.Sprintf("%s-%s-%s-%s.%s", namespace, name, provider, version, archiveFormat)
	return path.Join(modulePathPrefix(prefix, namespace, name, provider), f)
}

func readSHASums(r io.Reader, name string) (string, error) {
	scanner := bufio.NewScanner(r)

	sha := ""
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		if len(parts) != 3 {
			continue
		}

		if parts[2] == name {
			sha = parts[0]
			break
		}
	}

	if sha == "" {
		return "", fmt.Errorf("did not find package: %s in shasums file", name)
	}

	return sha, nil
}

func moduleFromObject(key string, fileExtension string) (*core.Module, error) {
	dir, file := path.Split(key)

	dirParts := strings.Split(dir, "/")
	for _, part := range dirParts {
		dirParts = dirParts[1:] // Remove the first item
		if part == string(internalModuleType) {
			break
		}
	}
	if len(dirParts) < 3 {
		return nil, fmt.Errorf("module key is invalid: expected 3 directory parts, but was %d", len(dirParts))
	}

	fileExtension = fmt.Sprintf(".%s", fileExtension) // Add the dot to the file extension
	if !strings.HasSuffix(file, fileExtension) {
		return nil, fmt.Errorf("expected file extension \"%s\" but found \"%s\"", fileExtension, path.Ext(file))
	}
	file = strings.TrimSuffix(file, fileExtension) // Remove the file extension

	filePrefix := fmt.Sprintf("%s-%s-%s-", dirParts[0], dirParts[1], dirParts[2])
	if !strings.HasPrefix(file, filePrefix) {
		return nil, fmt.Errorf("expected file prefix \"%s\" but file is \"%s\"", filePrefix, file)
	}
	version := strings.TrimPrefix(file, filePrefix) // Remove everything up to the version
	if version == "" {
		return nil, fmt.Errorf("module key is invalid, could not parse version")
	}

	return &core.Module{
		Namespace: dirParts[0],
		Name:      dirParts[1],
		Provider:  dirParts[2],
		Version:   version,
	}, nil
}
