package config

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewNoCacheDriverWithPureGoSQLite(t *testing.T) {
	conf := DatabaseConf{
		Type:        "sqlite3",
		DBPath:      filepath.Join(t.TempDir(), "simple-admin.db"),
		MaxOpenConn: 1,
	}
	driver := conf.NewNoCacheDriver()
	t.Cleanup(func() {
		require.NoError(t, driver.Close())
	})

	require.Equal(t, "sqlite", conf.sqlDriverName())
	require.Equal(t, "sqlite3", driver.Dialect())
	_, err := driver.DB().ExecContext(context.Background(), "CREATE TABLE test_pure_go_sqlite (id INTEGER PRIMARY KEY)")
	require.NoError(t, err)
}
