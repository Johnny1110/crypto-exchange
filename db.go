package main

import (
	"database/sql"
	"fmt"
	"github.com/labstack/gommon/log"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

// initDB if testMode = true, everytime startup the app, it will rebuild database with schema and prepare mock data.
func initDB(testMode bool) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./app/exg.db")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Infof("Database initialized successfully")

	// Run SQL files on startup if testMode
	if testMode {
		if err := runSQLFilesWithTransaction(db); err != nil {
			return nil, fmt.Errorf("failed to run SQL files: %w", err)
		}
		log.Infof("DB schema and testing data initialized successfully")
	}

	return db, err
}
