package db

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/lib/pq"
	"github.com/mmorpg-template/backend/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test database configuration
const (
	testDBHost     = "localhost"
	testDBPort     = "5432"
	testDBUser     = "test"
	testDBPassword = "test"
	testDBName     = "mmorpg_test"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Connect to postgres database to create test database
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword)
	
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)
	
	// Drop and recreate test database
	db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", testDBName))
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err)
	db.Close()
	
	// Connect to test database
	testConnStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		testDBHost, testDBPort, testDBUser, testDBPassword, testDBName)
	
	testDB, err := sql.Open("postgres", testConnStr)
	require.NoError(t, err)
	
	return testDB
}

func TestMigrator_Up(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	log := logger.NewTest()
	migrator, err := NewMigrator(db, log)
	require.NoError(t, err)
	defer migrator.Close()
	
	// Run migrations
	err = migrator.Up()
	assert.NoError(t, err)
	
	// Verify tables exist
	tables := []string{"users", "sessions"}
	for _, table := range tables {
		var exists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`, table).Scan(&exists)
		
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}
	
	// Verify version
	version, dirty, err := migrator.Version()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(2), version) // We have 2 migrations
}

func TestMigrator_DownUp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	log := logger.NewTest()
	migrator, err := NewMigrator(db, log)
	require.NoError(t, err)
	defer migrator.Close()
	
	// Run all migrations
	err = migrator.Up()
	require.NoError(t, err)
	
	// Rollback one migration
	err = migrator.Down()
	assert.NoError(t, err)
	
	// Verify sessions table doesn't exist
	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sessions'
		)`).Scan(&exists)
	assert.NoError(t, err)
	assert.False(t, exists, "Sessions table should not exist after rollback")
	
	// Verify users table still exists
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)`).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Users table should still exist")
	
	// Run migrations again
	err = migrator.Up()
	assert.NoError(t, err)
	
	// Verify sessions table exists again
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sessions'
		)`).Scan(&exists)
	assert.NoError(t, err)
	assert.True(t, exists, "Sessions table should exist after re-running migrations")
}

func TestMigrator_MigrateToVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	log := logger.NewTest()
	migrator, err := NewMigrator(db, log)
	require.NoError(t, err)
	defer migrator.Close()
	
	// Migrate to version 1
	err = migrator.Migrate(1)
	assert.NoError(t, err)
	
	version, dirty, err := migrator.Version()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(1), version)
	
	// Verify only users table exists
	var usersExists, sessionsExists bool
	db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'users'
		)`).Scan(&usersExists)
	db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sessions'
		)`).Scan(&sessionsExists)
	
	assert.True(t, usersExists)
	assert.False(t, sessionsExists)
	
	// Migrate to version 2
	err = migrator.Migrate(2)
	assert.NoError(t, err)
	
	version, dirty, err = migrator.Version()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Equal(t, uint(2), version)
}

func TestValidateMigrations(t *testing.T) {
	err := ValidateMigrations()
	assert.NoError(t, err)
}