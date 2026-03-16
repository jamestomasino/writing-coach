package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Store) Migrate(ctx context.Context, migrationsDir string) error {
	if _, err := s.SQL.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		version := entry.Name()
		var exists int
		if err := s.SQL.QueryRowContext(ctx, `
			SELECT COUNT(1) FROM schema_migrations WHERE version = ?
		`, version).Scan(&exists); err != nil {
			return err
		}
		if exists > 0 {
			continue
		}

		path := filepath.Join(migrationsDir, version)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := s.SQL.ExecContext(ctx, string(sqlBytes)); err != nil {
			if strings.Contains(err.Error(), "duplicate column name") {
				// Allow additive migrations to be replay-safe on fresh databases
				// when the base schema already includes the later column.
			} else {
				return fmt.Errorf("apply migration %s: %w", version, err)
			}
		}
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO schema_migrations (version) VALUES (?)
		`, version); err != nil {
			return err
		}
	}

	return nil
}
