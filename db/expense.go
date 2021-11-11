package db

import (
	"database/sql"
	"github.com/maxwww/accounting_bot/types"
	"log"
	"time"
)

func AddExpense(db *sql.DB, expense string, amount float64, tm time.Time, userId int) bool {
	var exists bool
	row := db.QueryRow("SELECT EXISTS (SELECT id FROM expenses WHERE expense = $1 and amount = $2 and created_at = $3)", expense, amount, tm)
	err := row.Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Print(err)
		return false
	}

	if !exists {
		_, err = db.Exec(`
INSERT INTO expenses (expense, amount, created_at, user_id )
VALUES ($1, $2, $3, $4)`, expense, amount, tm, userId)
		if err != nil {
			log.Print(err)
			return false
		}
	} else {
		return false
	}

	return true
}

func GetCurrentExpenses(db *sql.DB) ([]types.Expense, error) {
	location, _ := time.LoadLocation("Europe/Kiev")
	tm := time.Now().In(location)
	currentYear, currentMonth, _ := tm.Date()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
	lastOfMonth := firstOfMonth.AddDate(0, 1, 0)
	rows, err := db.Query(`select expense, sum(amount) am
	from expenses
	where created_at >= $1
	 and created_at < $2
	group by expense
	order by am desc;`, firstOfMonth, lastOfMonth)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []types.Expense

	for rows.Next() {
		var expense types.Expense
		if err := rows.Scan(&expense.Expense, &expense.Amount); err != nil {
			return expenses, err
		}
		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		return expenses, err
	}
	return expenses, nil
}

func GetLastMonthExpenses(db *sql.DB) ([]types.Expense, error) {
	location, _ := time.LoadLocation("Europe/Kiev")
	tm := time.Now().In(location)
	currentYear, currentMonth, _ := tm.Date()
	lastOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, location)
	firstOfMonth := lastOfMonth.AddDate(0, -1, 0)
	rows, err := db.Query(`select expense, sum(amount) am
	from expenses
	where created_at >= $1
	 and created_at < $2
	group by expense
	order by am desc;`, firstOfMonth, lastOfMonth)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []types.Expense

	for rows.Next() {
		var expense types.Expense
		if err := rows.Scan(&expense.Expense, &expense.Amount); err != nil {
			return expenses, err
		}
		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		return expenses, err
	}
	return expenses, nil
}
