package storage

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MemoryRepository is an in-memory implementation of DefinitionRepository
type MemoryRepository struct {
	definitions map[string]*Definition // key format: "tenantID:key:version"
	mu          sync.RWMutex
}

// NewMemoryRepository creates a new in-memory repository
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		definitions: make(map[string]*Definition),
	}
}

// Deploy saves a new version of a definition
func (r *MemoryRepository) Deploy(ctx context.Context, def *Definition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find the next version
	nextVersion := 1
	for _, d := range r.definitions {
		if d.Key == def.Key && d.TenantID == def.TenantID && d.Version >= nextVersion {
			nextVersion = d.Version + 1
		}
	}

	// Generate ID and set metadata
	def.ID = uuid.New().String()
	def.Version = nextVersion
	def.CreatedAt = time.Now()
	def.Checksum = computeChecksum(def.Source)

	// Store
	storageKey := makeStorageKey(def.TenantID, def.Key, def.Version)
	r.definitions[storageKey] = def

	return nil
}

// GetByKey returns the latest version of a definition
func (r *MemoryRepository) GetByKey(ctx context.Context, key string, tenantID string) (*Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *Definition
	for _, d := range r.definitions {
		if d.Key == key && d.TenantID == tenantID {
			if latest == nil || d.Version > latest.Version {
				latest = d
			}
		}
	}

	if latest == nil {
		return nil, fmt.Errorf("definition not found: key=%s, tenantId=%s", key, tenantID)
	}

	return latest, nil
}

// GetByKeyAndVersion returns a specific version of a definition
func (r *MemoryRepository) GetByKeyAndVersion(ctx context.Context, key string, version int, tenantID string) (*Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	storageKey := makeStorageKey(tenantID, key, version)
	def, ok := r.definitions[storageKey]
	if !ok {
		return nil, fmt.Errorf("definition not found: key=%s, version=%d, tenantId=%s", key, version, tenantID)
	}

	return def, nil
}

// List returns a list of definitions matching the filter
func (r *MemoryRepository) List(ctx context.Context, filter *ListFilter) ([]*Definition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Collect matching definitions (latest version only)
	latestVersions := make(map[string]*Definition) // key: "tenantID:key"

	for _, d := range r.definitions {
		// Apply filters
		if filter != nil {
			if filter.Key != "" && d.Key != filter.Key {
				continue
			}
			if filter.TenantID != "" && d.TenantID != filter.TenantID {
				continue
			}
		}

		latestKey := fmt.Sprintf("%s:%s", d.TenantID, d.Key)
		if existing, ok := latestVersions[latestKey]; !ok || d.Version > existing.Version {
			latestVersions[latestKey] = d
		}
	}

	// Convert to slice and sort by name
	result := make([]*Definition, 0, len(latestVersions))
	for _, d := range latestVersions {
		result = append(result, d)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(result) {
			result = result[filter.Offset:]
		} else if filter.Offset >= len(result) {
			return []*Definition{}, nil
		}

		if filter.Limit > 0 && filter.Limit < len(result) {
			result = result[:filter.Limit]
		}
	}

	return result, nil
}

// Delete deletes all versions of a definition
func (r *MemoryRepository) Delete(ctx context.Context, key string, tenantID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	keysToDelete := make([]string, 0)
	for storageKey, d := range r.definitions {
		if d.Key == key && d.TenantID == tenantID {
			keysToDelete = append(keysToDelete, storageKey)
		}
	}

	if len(keysToDelete) == 0 {
		return fmt.Errorf("definition not found: key=%s, tenantId=%s", key, tenantID)
	}

	for _, k := range keysToDelete {
		delete(r.definitions, k)
	}

	return nil
}

func makeStorageKey(tenantID, key string, version int) string {
	return fmt.Sprintf("%s:%s:%d", tenantID, key, version)
}
