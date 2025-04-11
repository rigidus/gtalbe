package database

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/compose"
)

func TestNewDriver_WithDockerCompose(t *testing.T) {
	ctx := context.Background()
	_, currentFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "cannot get current file location")

	composeFile := "docker_compose.yml"
	composePath := filepath.Join(filepath.Dir(currentFile), "../../", composeFile)
	absComposePath, err := filepath.Abs(composePath)
	require.NoError(t, err, "failed to get absolute path to docker-compose.yml")

	compose, err := compose.NewDockerCompose(absComposePath)
	require.NoError(t, err)

	err = compose.Up(ctx)
	require.NoError(t, err)

	t.Cleanup(func() { compose.Down(ctx) })

	time.Sleep(5 * time.Second)

	dsn := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"

	logger := logrus.New()
	driver, err := NewDriver(logger, dsn)
	require.NoError(t, err)
	require.NotNil(t, driver)

	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	var exists bool
	query := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'transactions'
		);
	`
	err = db.QueryRow(query).Scan(&exists)
	require.NoError(t, err)
	require.True(t, exists, "transactions table should exist after migration")
}
