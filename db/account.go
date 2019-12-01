package db

import (
	"database/sql"
	"fmt"
)

func UpdateAccount(db *sql.DB, slug string, balance float64) error {
	sqlStatement := `
UPDATE accounts
SET balance = $1
WHERE slug = $2;`
	result, err := db.Exec(sqlStatement, balance, slug) // OK
	if err != nil {
		return err
	}
	rowCnt, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowCnt != 1 {
		return fmt.Errorf("more one rows were affected")
	}

	return nil
}
