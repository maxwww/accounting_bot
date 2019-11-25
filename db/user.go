package db

import (
	"database/sql"
	"log"
)

func LogUser(db *sql.DB, userId int, isBot bool, firstName string, lastName string, userName string, languageCode string) {
	var exists bool
	row := db.QueryRow("SELECT EXISTS (SELECT id FROM users WHERE id = $1)", userId)
	err := row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
	}
	if !exists {
		_, err = db.Exec(`
INSERT INTO users (id, is_bot, first_name, last_name, user_name, language_code, requests )
VALUES ($1, $2, $3, $4, $5, $6, $7)`, userId, isBot, firstName, lastName, userName, languageCode, 1)
		if err != nil {
			log.Print(err)
		}
	} else {
		_, err = db.Exec(`
UPDATE users
SET requests = requests + 1
WHERE id = $1;`, userId)
		if err != nil {
			log.Print(err)
		}
	}
}
