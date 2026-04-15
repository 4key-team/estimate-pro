package testutil

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// SetupPostgres starts a PostgreSQL container and returns a connected pool.
// Migrations are applied automatically.
func SetupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := t.Context()

	ctr, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	if err != nil {
		t.Fatalf("testutil.SetupPostgres: start container: %v", err)
	}
	t.Cleanup(func() { ctr.Terminate(context.Background()) })

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("testutil.SetupPostgres: connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("testutil.SetupPostgres: new pool: %v", err)
	}
	t.Cleanup(func() { pool.Close() })

	runMigrations(t, pool)
	return pool
}

// SetupRedis starts a Redis container and returns a connected client.
func SetupRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := t.Context()

	ctr, err := tcredis.Run(ctx, "redis:8-alpine")
	if err != nil {
		t.Fatalf("testutil.SetupRedis: start container: %v", err)
	}
	t.Cleanup(func() { ctr.Terminate(context.Background()) })

	uri, err := ctr.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("testutil.SetupRedis: connection string: %v", err)
	}

	opts, err := redis.ParseURL(uri)
	if err != nil {
		t.Fatalf("testutil.SetupRedis: parse URL: %v", err)
	}

	client := redis.NewClient(opts)
	t.Cleanup(func() { client.Close() })

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("testutil.SetupRedis: ping: %v", err)
	}
	return client
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, thisFile, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations")

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("testutil.runMigrations: read dir: %v", err)
	}

	// Sort to apply in order
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			t.Fatalf("testutil.runMigrations: read %s: %v", entry.Name(), err)
		}

		// Extract only the Up section (before -- +goose Down)
		sql := string(data)
		if idx := indexOf(sql, "-- +goose Down"); idx >= 0 {
			sql = sql[:idx]
		}
		// Remove goose directive
		sql = replaceAll(sql, "-- +goose Up", "")

		if _, err := pool.Exec(t.Context(), sql); err != nil {
			t.Fatalf("testutil.runMigrations: exec %s: %v", entry.Name(), err)
		}
	}

	// Columns not in migration files but used by code
	extras := []string{
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(512) DEFAULT ''`,
	}
	for _, ddl := range extras {
		if _, err := pool.Exec(t.Context(), ddl); err != nil {
			t.Fatalf("testutil.runMigrations: extra DDL: %v", err)
		}
	}
}

func indexOf(s, substr string) int {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		i := indexOf(s, old)
		if i < 0 {
			return result + s
		}
		result += s[:i] + new
		s = s[i+len(old):]
	}
}
