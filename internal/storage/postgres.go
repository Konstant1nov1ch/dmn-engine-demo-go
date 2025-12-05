package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRepository реализует DefinitionRepository с PostgreSQL
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository создаёт новый PostgreSQL репозиторий
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// Deploy сохраняет новую версию definition
func (r *PostgresRepository) Deploy(ctx context.Context, def *Definition) error {
	// Получаем следующую версию
	var nextVersion int
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(version), 0) + 1 
		FROM dmn_definitions 
		WHERE key = $1 AND (
			(tenant_id = $2) OR 
			($2 IS NULL AND tenant_id IS NULL) OR
			($2 = '' AND tenant_id IS NULL)
		)
	`, def.Key, nullableString(def.TenantID)).Scan(&nextVersion)
	if err != nil {
		return fmt.Errorf("failed to get next version: %w", err)
	}

	// Генерируем ID и метаданные
	def.ID = uuid.New().String()
	def.Version = nextVersion
	def.CreatedAt = time.Now()
	def.Checksum = computeChecksum(def.Source)

	// Сериализуем parsed model в JSON
	parsedJSON, err := json.Marshal(def.ParsedModel)
	if err != nil {
		return fmt.Errorf("failed to marshal parsed model: %w", err)
	}

	// Сохраняем
	_, err = r.pool.Exec(ctx, `
		INSERT INTO dmn_definitions (id, key, version, name, source, parsed_model, checksum, tenant_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, def.ID, def.Key, def.Version, def.Name, def.Source, parsedJSON, def.Checksum, nullableString(def.TenantID), def.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert definition: %w", err)
	}

	return nil
}

// GetByKey возвращает последнюю версию definition
func (r *PostgresRepository) GetByKey(ctx context.Context, key string, tenantID string) (*Definition, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, key, version, name, source, parsed_model, checksum, tenant_id, created_at
		FROM dmn_definitions
		WHERE key = $1 AND (
			(tenant_id = $2) OR 
			($2 IS NULL AND tenant_id IS NULL) OR
			($2 = '' AND tenant_id IS NULL)
		)
		ORDER BY version DESC
		LIMIT 1
	`, key, nullableString(tenantID))

	return r.scanDefinition(row)
}

// GetByKeyAndVersion возвращает конкретную версию definition
func (r *PostgresRepository) GetByKeyAndVersion(ctx context.Context, key string, version int, tenantID string) (*Definition, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, key, version, name, source, parsed_model, checksum, tenant_id, created_at
		FROM dmn_definitions
		WHERE key = $1 AND version = $2 AND (
			(tenant_id = $3) OR 
			($3 IS NULL AND tenant_id IS NULL) OR
			($3 = '' AND tenant_id IS NULL)
		)
	`, key, version, nullableString(tenantID))

	return r.scanDefinition(row)
}

// List возвращает список definitions
func (r *PostgresRepository) List(ctx context.Context, filter *ListFilter) ([]*Definition, error) {
	// Строим запрос с latest версией для каждого key
	query := `
		SELECT DISTINCT ON (key, tenant_id) 
			id, key, version, name, source, parsed_model, checksum, tenant_id, created_at
		FROM dmn_definitions
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter != nil {
		if filter.Key != "" {
			query += fmt.Sprintf(" AND key = $%d", argNum)
			args = append(args, filter.Key)
			argNum++
		}
		if filter.TenantID != "" {
			query += fmt.Sprintf(" AND tenant_id = $%d", argNum)
			args = append(args, filter.TenantID)
			argNum++
		}
	}

	query += " ORDER BY key, tenant_id, version DESC"

	if filter != nil && filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter != nil && filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query definitions: %w", err)
	}
	defer rows.Close()

	var definitions []*Definition
	for rows.Next() {
		def, err := r.scanDefinitionFromRows(rows)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, def)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return definitions, nil
}

// Delete удаляет все версии definition
func (r *PostgresRepository) Delete(ctx context.Context, key string, tenantID string) error {
	result, err := r.pool.Exec(ctx, `
		DELETE FROM dmn_definitions
		WHERE key = $1 AND (
			(tenant_id = $2) OR 
			($2 IS NULL AND tenant_id IS NULL) OR
			($2 = '' AND tenant_id IS NULL)
		)
	`, key, nullableString(tenantID))
	if err != nil {
		return fmt.Errorf("failed to delete definition: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("definition not found: key=%s, tenantId=%s", key, tenantID)
	}

	return nil
}

// GetAllVersions возвращает все версии definition
func (r *PostgresRepository) GetAllVersions(ctx context.Context, key string, tenantID string) ([]*Definition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, key, version, name, source, parsed_model, checksum, tenant_id, created_at
		FROM dmn_definitions
		WHERE key = $1 AND (tenant_id = $2 OR ($2 = '' AND tenant_id IS NULL))
		ORDER BY version DESC
	`, key, nullableString(tenantID))
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var definitions []*Definition
	for rows.Next() {
		def, err := r.scanDefinitionFromRows(rows)
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, def)
	}

	return definitions, nil
}

// scanDefinition сканирует одну строку в Definition
func (r *PostgresRepository) scanDefinition(row pgx.Row) (*Definition, error) {
	var def Definition
	var parsedJSON []byte
	var tenantID *string

	err := row.Scan(
		&def.ID,
		&def.Key,
		&def.Version,
		&def.Name,
		&def.Source,
		&parsedJSON,
		&def.Checksum,
		&tenantID,
		&def.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("definition not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan definition: %w", err)
	}

	if tenantID != nil {
		def.TenantID = *tenantID
	}

	// Десериализуем parsed model
	if len(parsedJSON) > 0 {
		if err := json.Unmarshal(parsedJSON, &def.ParsedModel); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parsed model: %w", err)
		}
	}

	return &def, nil
}

// scanDefinitionFromRows сканирует строку из rows
func (r *PostgresRepository) scanDefinitionFromRows(rows pgx.Rows) (*Definition, error) {
	var def Definition
	var parsedJSON []byte
	var tenantID *string

	err := rows.Scan(
		&def.ID,
		&def.Key,
		&def.Version,
		&def.Name,
		&def.Source,
		&parsedJSON,
		&def.Checksum,
		&tenantID,
		&def.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan definition: %w", err)
	}

	if tenantID != nil {
		def.TenantID = *tenantID
	}

	if len(parsedJSON) > 0 {
		if err := json.Unmarshal(parsedJSON, &def.ParsedModel); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parsed model: %w", err)
		}
	}

	return &def, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func computeChecksum(source string) string {
	h := sha256.Sum256([]byte(source))
	return hex.EncodeToString(h[:])
}
