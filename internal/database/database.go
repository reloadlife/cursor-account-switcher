package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("state database not found: %s", path)
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
		return "", fmt.Errorf("ItemTable not found in state database")
	}
	return name, nil
}

func ReadKeys(db *sql.DB, prefix string, patterns []string) (map[string]string, error) {
	table, err := resolveItemTable(db)
	if err != nil {
		return nil, err
	}

	keys := make(map[string]string)
	seen := make(map[string]bool)

	if prefix != "" {
		rows, err := db.Query(fmt.Sprintf(`SELECT key, value FROM "%s" WHERE key LIKE ?`, table), prefix+"%")
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				rows.Close()
				return nil, err
			}
			keys[key] = value
			seen[key] = true
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	for _, pattern := range patterns {
		rows, err := db.Query(fmt.Sprintf(`SELECT key, value FROM "%s" WHERE key LIKE ?`, table), pattern)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				rows.Close()
				return nil, err
			}
			if !seen[key] {
				keys[key] = value
				seen[key] = true
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	return keys, nil
}

func ReadKey(db *sql.DB, key string) (string, error) {
	table, err := resolveItemTable(db)
	if err != nil {
		return "", err
	}

	var value string
	err = db.QueryRow(
		fmt.Sprintf(`SELECT value FROM "%s" WHERE key = ? LIMIT 1`, table),
		key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return value, nil
}

func WriteKeys(db *sql.DB, prefix string, patterns []string, keys map[string]string) error {
	table, err := resolveItemTable(db)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if prefix != "" {
		if _, err := tx.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE key LIKE ?`, table), prefix+"%"); err != nil {
			return err
		}
	}

	for _, pattern := range patterns {
		if _, err := tx.Exec(fmt.Sprintf(`DELETE FROM "%s" WHERE key LIKE ?`, table), pattern); err != nil {
			return err
		}
	}

	stmt, err := tx.Prepare(fmt.Sprintf(`INSERT INTO "%s" (key, value) VALUES (?, ?)`, table))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for key, value := range keys {
		if !matchesManagedKey(key, prefix, patterns) {
			continue
		}
		if _, err := stmt.Exec(key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func matchesManagedKey(key, prefix string, patterns []string) bool {
	if prefix != "" && strings.HasPrefix(key, prefix) {
		return true
	}
	for _, pattern := range patterns {
		if likeMatch(key, pattern) {
			return true
		}
	}
	return false
}

func likeMatch(key, pattern string) bool {
	if !strings.Contains(pattern, "%") {
		return key == pattern
	}
	parts := strings.Split(pattern, "%")
	if !strings.HasPrefix(key, parts[0]) {
		return false
	}
	key = key[len(parts[0]):]
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(key, parts[i])
		if idx < 0 {
			return false
		}
		key = key[idx+len(parts[i]):]
	}
	return strings.HasSuffix(key, parts[len(parts)-1])
}

func ReadAuthKeys(db *sql.DB) (map[string]string, error) {
	return ReadKeys(db, "cursorAuth/", nil)
}

func ReadCurrentEmail(db *sql.DB) (string, error) {
	return ReadKey(db, "cursorAuth/cachedEmail")
}

func WriteAuthKeys(db *sql.DB, keys map[string]string) error {
	return WriteKeys(db, "cursorAuth/", nil, keys)
}
