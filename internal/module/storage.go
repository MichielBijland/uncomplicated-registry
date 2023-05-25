package module

import (
	"context"
	"io"

	"github.com/MichielBijland/uncomplicated-registry/internal/core"
)

// Storage represents the repository of Terraform modules.
type Storage interface {
	GetModule(ctx context.Context, namespace, name, provider, version string) (core.Module, error)
	ListModuleVersions(ctx context.Context, namespace, name, provider string) ([]core.Module, error)
	UploadModule(ctx context.Context, namespace, name, provider, version string, body io.Reader) (core.Module, error)
}
