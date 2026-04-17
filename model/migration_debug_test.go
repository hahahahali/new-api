package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestDebugSQLiteMigrationsAgainstLocalDB(t *testing.T) {
	if os.Getenv("RUN_LOCAL_SQLITE_MIGRATION_DEBUG") == "" {
		t.Skip("set RUN_LOCAL_SQLITE_MIGRATION_DEBUG=1 to run the local sqlite migration debug test")
	}

	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	initCol()

	tmpDBPath := filepath.Join(t.TempDir(), "one-api.db")
	rawDB, err := os.ReadFile("../one-api.db")
	if err != nil {
		t.Fatalf("read source db: %v", err)
	}
	if err := os.WriteFile(tmpDBPath, rawDB, 0o600); err != nil {
		t.Fatalf("write temp db: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(tmpDBPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	DB = db
	if err := migrateDB(); err != nil {
		t.Fatalf("migrateDB failed: %v", err)
	}
}
