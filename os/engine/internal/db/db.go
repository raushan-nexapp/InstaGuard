// Package db handles NexappOS's SQLite configuration database.
//
// This is NexappOS's equivalent of FortiGate's CMDB (Configuration
// Management Database). All firewall config lives here — interfaces,
// policies, addresses, services, DHCP settings, DNS settings.
//
// The engine reads from this database and generates the real config
// files on disk (/etc/nftables.conf, /etc/dhcp/dhcpd.conf, etc).
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a SQL database connection.
type DB struct {
	*sql.DB
}

// Open creates or opens the NexappOS database at the given path.
// If the file does not exist, it will be created and the schema applied.
func Open(path string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	db := &DB{DB: conn}

	if err := db.applySchema(); err != nil {
		return nil, fmt.Errorf("apply schema: %w", err)
	}

	log.Printf("database ready at %s", path)
	return db, nil
}

// applySchema creates all tables if they don't already exist.
// Safe to run on every startup.
func (db *DB) applySchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS interfaces (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		name        TEXT    NOT NULL UNIQUE,
		role        TEXT    NOT NULL CHECK(role IN ('wan','lan','dmz','mgmt')),
		ipv4_cidr   TEXT,
		description TEXT,
		enabled     INTEGER NOT NULL DEFAULT 1,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS policies (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT    NOT NULL UNIQUE,
		src_interface TEXT    NOT NULL,
		dst_interface TEXT    NOT NULL,
		src_address   TEXT    NOT NULL DEFAULT 'any',
		dst_address   TEXT    NOT NULL DEFAULT 'any',
		protocol      TEXT    NOT NULL DEFAULT 'any',
		dst_port      TEXT    NOT NULL DEFAULT 'any',
		action        TEXT    NOT NULL CHECK(action IN ('accept','drop','reject')),
		enabled       INTEGER NOT NULL DEFAULT 1,
		priority      INTEGER NOT NULL DEFAULT 100,
		description   TEXT,
		created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_policies_priority ON policies(priority);
	CREATE INDEX IF NOT EXISTS idx_policies_enabled  ON policies(enabled);
	`

	_, err := db.Exec(schema)
	return err
}
