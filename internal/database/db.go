package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"chat-system-go/internal/config"
)

var DB *sql.DB

func Init() error {
	dbPath := config.AppConfig.DBPath
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_foreign_keys=on")
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	DB = db

	if err := createTables(); err != nil {
		return err
	}
	log.Println("SQLite database initialized")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS sessions (
		sid TEXT PRIMARY KEY,
		token TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		last_active INTEGER NOT NULL,
		expire_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS messages (
		id TEXT PRIMARY KEY,
		sid TEXT NOT NULL REFERENCES sessions(sid) ON DELETE CASCADE,
		role TEXT NOT NULL,
		type TEXT NOT NULL,
		text TEXT,
		file_name TEXT,
		file_size INTEGER,
		file_url TEXT,
		mime_type TEXT,
		time INTEGER NOT NULL,
		metadata TEXT
	);
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		sid TEXT NOT NULL REFERENCES sessions(sid) ON DELETE CASCADE,
		path TEXT NOT NULL,
		size INTEGER NOT NULL,
		created_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS topics (
		sid TEXT PRIMARY KEY REFERENCES sessions(sid) ON DELETE CASCADE,
		topic_id INTEGER NOT NULL,
		topic_name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS device_names (
		sid TEXT PRIMARY KEY REFERENCES sessions(sid) ON DELETE CASCADE,
		device_type TEXT NOT NULL,
		seq TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rate_limits (
		ip TEXT NOT NULL,
		window_start INTEGER NOT NULL,
		count INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		PRIMARY KEY (ip, window_start)
	);
	CREATE TABLE IF NOT EXISTS blocked_ips (
		ip TEXT PRIMARY KEY,
		blocked_at INTEGER NOT NULL,
		expire_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS autoreply_flags (
		sid TEXT PRIMARY KEY REFERENCES sessions(sid) ON DELETE CASCADE,
		sent_at INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS global_seq (
		key TEXT PRIMARY KEY,
		value INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_messages_sid_time ON messages(sid, time);
	CREATE INDEX IF NOT EXISTS idx_sessions_expire ON sessions(expire_at);
	CREATE INDEX IF NOT EXISTS idx_files_sid ON files(sid);
	CREATE INDEX IF NOT EXISTS idx_blocked_ips_expire ON blocked_ips(expire_at);
	`
	_, err := DB.Exec(schema)
	return err
}
