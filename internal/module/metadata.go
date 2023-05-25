package module

import (
	"errors"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
)

// Metadata provides information about a given module version.
type Metadata struct {
	Namespace string `hcl:"namespace" json:"namespace"`
	Name      string `hcl:"name" json:"name"`
	Provider  string `hcl:"provider" json:"provider"`
	Version   string `hcl:"version" json:"version"`
}

// Validate ensures that a Metadata is valid.
func (m *Metadata) Validate() error {
	var result *multierror.Error

	if m.Namespace == "" {
		result = multierror.Append(result, errors.New("metadata.namespace cannot be empty"))
	}

	if m.Name == "" {
		result = multierror.Append(result, errors.New("metadata.name cannot be empty"))
	}

	if m.Provider == "" {
		result = multierror.Append(result, errors.New("metadata.provider cannot be empty"))
	}

	if _, err := version.NewVersion(m.Version); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

func (m *Metadata) String() string {
	return fmt.Sprintf("%s/%s/%s/%s", m.Namespace, m.Name, m.Provider, m.Version)
}
