package database

import (
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRunDataMigrationsRunsEachMigrationOnce(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "finance.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	require.NoError(t, err)

	runs := 0
	migrations := []DataMigration{{
		ID: "test_001",
		Up: func(tx *gorm.DB) error {
			runs++
			return tx.Exec("CREATE TABLE IF NOT EXISTS migration_test_values (id INTEGER PRIMARY KEY)").Error
		},
	}}

	require.NoError(t, RunDataMigrations(db, dbPath, migrations))
	require.Equal(t, 1, runs)
	require.DirExists(t, filepath.Join(filepath.Dir(dbPath), "backups"))

	require.NoError(t, RunDataMigrations(db, dbPath, migrations))
	require.Equal(t, 1, runs)
}
