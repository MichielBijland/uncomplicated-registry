package storage

import "errors"

// Storage errors.
var (
	// module errors
	ErrModuleUploadFailed  = errors.New("failed to upload module")
	ErrModuleAlreadyExists = errors.New("module already exists")
	ErrModuleNotFound      = errors.New("failed to locate module")
	ErrModuleListFailed    = errors.New("failed to list module versions")
)
