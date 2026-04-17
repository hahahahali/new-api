package model

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const legacySQLiteUsersDDL = `
CREATE TABLE users (
  id integer,
  username text UNIQUE,
  password text NOT NULL,
  display_name text,
  role integer DEFAULT 1,
  status integer DEFAULT 1,
  email text,
  github_id text,
  discord_id text,
  oidc_id text,
  wechat_id text,
  telegram_id text,
  access_token char(32),
  quota integer DEFAULT 0,
  used_quota integer DEFAULT 0,
  request_count integer DEFAULT 0,
  "group" varchar(64) DEFAULT 'default',
  aff_code varchar(32),
  aff_count integer DEFAULT 0,
  aff_quota integer DEFAULT 0,
  aff_history integer DEFAULT 0,
  inviter_id integer,
  deleted_at datetime,
  linux_do_id text,
  setting text,
  remark varchar(255),
  stripe_customer varchar(64),
  PRIMARY KEY (id)
);
CREATE INDEX idx_users_display_name ON users(display_name);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_github_id ON users(github_id);
CREATE INDEX idx_users_discord_id ON users(discord_id);
CREATE INDEX idx_users_oidc_id ON users(oidc_id);
CREATE INDEX idx_users_wechat_id ON users(wechat_id);
CREATE INDEX idx_users_telegram_id ON users(telegram_id);
CREATE INDEX idx_users_inviter_id ON users(inviter_id);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_linux_do_id ON users(linux_do_id);
CREATE INDEX idx_users_stripe_customer ON users(stripe_customer);
`

func openSQLiteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "migration.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func seedLegacyUsersTable(t *testing.T, db *gorm.DB) {
	t.Helper()

	if err := db.Exec(legacySQLiteUsersDDL).Error; err != nil {
		t.Fatalf("create legacy users table: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (
		id, username, password, display_name, role, status, quota, used_quota, request_count, "group", aff_code, aff_count, aff_quota, aff_history, setting
	) VALUES (
		1, 'u1', '$2a$10$fZb/.Fe9RVnzpmw.hWOJ1u4DIpR4uPEBH7Pd3TIiBiX5scxIkfrau', 'User 1', 100, 1, 100, 0, 0, 'default', 'ABCD', 0, 0, 0, '{"k":1}'
	)`).Error; err != nil {
		t.Fatalf("seed legacy users row: %v", err)
	}
}

func withSQLiteMigrationTestEnv(t *testing.T, db *gorm.DB) {
	t.Helper()

	oldDB := DB
	oldSQLite := common.UsingSQLite
	oldMySQL := common.UsingMySQL
	oldPostgres := common.UsingPostgreSQL

	DB = db
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	initCol()

	t.Cleanup(func() {
		DB = oldDB
		common.UsingSQLite = oldSQLite
		common.UsingMySQL = oldMySQL
		common.UsingPostgreSQL = oldPostgres
	})
}

func countUsers(t *testing.T, db *gorm.DB) int64 {
	t.Helper()

	var count int64
	if err := db.Table("users").Count(&count).Error; err != nil {
		t.Fatalf("count users: %v", err)
	}
	return count
}

func TestEnsureUserTableSQLitePreservesLegacyRows(t *testing.T) {
	db := openSQLiteTestDB(t)
	seedLegacyUsersTable(t, db)
	withSQLiteMigrationTestEnv(t, db)

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("unexpected seed count: got %d want 1", got)
	}

	if err := ensureUserTableSQLite(); err != nil {
		t.Fatalf("ensureUserTableSQLite failed: %v", err)
	}

	if got := countUsers(t, db); got != 1 {
		t.Fatalf("legacy users row was lost after safe sqlite migration: got %d want 1", got)
	}

	var user User
	if err := db.First(&user, 1).Error; err != nil {
		t.Fatalf("load migrated user: %v", err)
	}
	if user.Username != "u1" || user.AffCode != "ABCD" || user.Group != "default" {
		t.Fatalf("unexpected migrated user: %+v", user)
	}
}

func TestSQLiteAutoMigrateUserRecreatesLegacyTable(t *testing.T) {
	if os.Getenv("RUN_UNSAFE_SQLITE_AUTOMIGRATE_REPRO") == "" {
		t.Skip("set RUN_UNSAFE_SQLITE_AUTOMIGRATE_REPRO=1 to reproduce the legacy unsafe sqlite AutoMigrate path")
	}

	db := openSQLiteTestDB(t)
	seedLegacyUsersTable(t, db)
	withSQLiteMigrationTestEnv(t, db)

	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("sqlite AutoMigrate user failed: %v", err)
	}

	t.Logf("user count after sqlite AutoMigrate: %d", countUsers(t, db))
}
