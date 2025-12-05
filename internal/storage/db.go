package storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/dmn-engine-go/internal/config"
)

// NewPool создаёт пул соединений с PostgreSQL из конфигурации
func NewPool(ctx context.Context, cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Применяем настройки пула
	poolConfig.MaxConns = cfg.MaxConns
	poolConfig.MinConns = cfg.MinConns
	poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// RunMigrations выполняет миграции БД
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS dmn_definitions (
			id          UUID PRIMARY KEY,
			key         VARCHAR(255) NOT NULL,
			version     INT NOT NULL DEFAULT 1,
			name        VARCHAR(255),
			source      TEXT NOT NULL,
			parsed_model JSONB,
			checksum    VARCHAR(64),
			tenant_id   VARCHAR(64),
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			
			CONSTRAINT unique_definition UNIQUE(key, version, tenant_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_dmn_def_key ON dmn_definitions(key)`,
		`CREATE INDEX IF NOT EXISTS idx_dmn_def_tenant ON dmn_definitions(tenant_id) WHERE tenant_id IS NOT NULL`,
		`CREATE INDEX IF NOT EXISTS idx_dmn_def_key_version ON dmn_definitions(key, version DESC)`,
	}

	for _, migration := range migrations {
		if _, err := pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
