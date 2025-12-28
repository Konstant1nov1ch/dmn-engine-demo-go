package storage

import (
	"context"
	"time"

	"github.com/konstantin/dmn-engine-go/internal/dmn"
)

// Definition represents a stored DMN definition
type Definition struct {
	ID          string           `json:"id"`
	Key         string           `json:"key"`
	Version     int              `json:"version"`
	Name        string           `json:"name"`
	Source      string           `json:"source"`      // Original XML
	ParsedModel *dmn.Definitions `json:"parsedModel"` // Parsed DMN model
	Checksum    string           `json:"checksum"`    // SHA256 of source
	TenantID    string           `json:"tenantId,omitempty"`
	CreatedAt   time.Time        `json:"createdAt"`
}

// ListFilter is used to filter definitions
type ListFilter struct {
	Key      string
	TenantID string
	Limit    int
	Offset   int
}

// DefinitionRepository is the interface for definition storage
type DefinitionRepository interface {
	// Deploy saves a new version of a definition
	Deploy(ctx context.Context, def *Definition) error

	// GetByKey returns the latest version of a definition
	GetByKey(ctx context.Context, key string, tenantID string) (*Definition, error)

	// GetByKeyAndVersion returns a specific version of a definition
	GetByKeyAndVersion(ctx context.Context, key string, version int, tenantID string) (*Definition, error)

	// List returns a list of definitions matching the filter
	List(ctx context.Context, filter *ListFilter) ([]*Definition, error)

	// Delete deletes all versions of a definition
	Delete(ctx context.Context, key string, tenantID string) error
}

