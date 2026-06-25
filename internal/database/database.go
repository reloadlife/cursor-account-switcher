package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const authPrefix = "cursorAuth/"

func Open(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("cursor state database not found: %s", path)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func resolveItemTable(db *sql.DB) (string, error) {
	var name string
	err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type='table' AND lower(name)='itemtable' LIMIT 1",
	).Scan(&name)
	if err != nil {
		return "", fmt.Errorf("ItemTable not found in Cursor state database")
	}
	return name, nil
}

func ReadAuthKeys(db *sql.DB) (map[string]string, error) {
	table, err := resolveItemTable(db)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(fmt.Sprintf(`SELECT key, value FROM "%s" WHERE key LIKE ?`, table), authPrefix+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		keys[key] = value
	}

	return keys, rows.Err()
}

func ReadCurrentEmail(db *sql.DB) (string, error) {
	table, err := resolveItemTable(db)
	if err != nil {
		return "", err
	}

	var email string
	err = db.QueryRow(
		fmt.Sprintf(`SELECT value FROM "%s" WHERE key = 'cursorAuth/cachedEmail' LIMIT 1`, table),
	).Scan(&email)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return email, nil
}

func WriteAuthKeys(db *sql.DB, keys map[string]string) error {
	table, err := resolveItemTable(db)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE key LIKE ?`, table), authPrefix+"%"); err != nil {
		return err
	}

	stmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO "%s" (key, value) VALUES (?, ?)`, table))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range keys {
		if _, err := stmt.Exec(key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}
