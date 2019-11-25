package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func NewConnection() *sql.DB {
	pgUser := os.Getenv("PG_USER")
	pgBasename := os.Getenv("PG_BASENAME")
	pgPassword := os.Getenv("PG_PASSWORD")
	pgHost := os.Getenv("PG_HOST")

	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable", pgHost, pgBasename, pgUser, pgPassword)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error connect to DB", err)
	}

	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users
	(
		id  int UNIQUE PRIMARY KEY,
		is_bot BOOLEAN NOT NULL,
		first_name VARCHAR(250) NOT NULL,
		last_name VARCHAR(250),
		user_name VARCHAR(250),
		language_code VARCHAR(250),
		requests int NOT NULL
	)
`)
	if err != nil {
		log.Fatal("Error to create table", err)
	}
	return db
}
