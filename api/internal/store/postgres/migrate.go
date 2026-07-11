package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies all pending SQL migrations from the given directory.
// It uses a simple version-based tracking system with a schema_migrations table.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	slog.Info("running database migrations", "dir", migrationsDir)

	// Ensure migrations table exists
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir %s: %w", migrationsDir, err)
	}

	// Collect up migration files
	type upMigration struct {
		version int
		path    string
	}

	var upMigrations []upMigration
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".up.sql") {
			continue
		}

		var version int
		if _, err := fmt.Sscanf(name, "%06d_", &version); err != nil {
			continue
		}

		upMigrations = append(upMigrations, upMigration{
			version: version,
			path:    filepath.Join(migrationsDir, name),
		})
	}

	sort.Slice(upMigrations, func(i, j int) bool {
		return upMigrations[i].version < upMigrations[j].version
	})

	// Get currently applied versions
	appliedVersions, err := getAppliedVersions(ctx, pool)
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	appliedSet := make(map[int]bool)
	for _, v := range appliedVersions {
		appliedSet[v] = true
	}

	// Apply pending migrations
	pendingCount := 0
	for _, m := range upMigrations {
		if appliedSet[m.version] {
			slog.Debug("migration already applied", "version", m.version)
			continue
		}

		sql, err := os.ReadFile(m.path)
		if err != nil {
			return fmt.Errorf("read migration %d: %w", m.version, err)
		}

		slog.Info("applying migration", "version", m.version, "path", m.path)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
		}

		_, err = tx.Exec(ctx, string(sql))
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("apply migration %d: %w", m.version, err)
		}

		_, err = tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, m.version)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}

		pendingCount++
		slog.Info("migration applied successfully", "version", m.version)
	}

	slog.Info("database migrations complete",
		"total", len(upMigrations),
		"applied", pendingCount,
	)
	return nil
}

// getAppliedVersions returns the list of already-applied migration versions.
func getAppliedVersions(ctx context.Context, pool *pgxpool.Pool) ([]int, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []int
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, nil
}

// MustMigrate runs migrations and panics on failure (for use in main()).
func MustMigrate(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) {
	if err := RunMigrations(ctx, pool, migrationsDir); err != nil {
		slog.Error("migration failed", "error", err)
		panic(fmt.Sprintf("database migration failed: %v", err))
	}
}

// DefaultMigrationsDir is the default path to migration files relative to the project root.
const DefaultMigrationsDir = "db/migrations"

// ErrNoMigrationsFound is returned when no .up.sql files exist in the migrations directory.
var ErrNoMigrationsFound = errors.New("no migration files found")
