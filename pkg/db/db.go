package db

import (
	"database/sql"
	_ "modernc.org/sqlite"
	"os"
)

var db *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(255) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init выполняет инициализацию SQLite базы данных для приложения
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	db, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		return err
	}
	if install {
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	}

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
