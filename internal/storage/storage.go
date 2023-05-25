package storage

import (
	"github.com/MichielBijland/uncomplicated-registry/internal/module"
)

const (
	DefaultModuleArchiveFormat = "tar.gz"
)

type Storage interface {
	module.Storage
}
