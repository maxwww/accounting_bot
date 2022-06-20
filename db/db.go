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

	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable TimeZone=Europe/Kiev", pgHost, pgBasename, pgUser, pgPassword)
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
		log.Fatal("Error to create table users", err)
	}

	_, err = db.Exec(`
create table if not exists accounts
(
	id serial not null
		constraint accounts_pkey
			primary key,
	slug varchar(255) not null,
	name varchar(255) not null,
	currency smallint not null,
	balance double precision default 0,
	priority integer
);

`)
	if err != nil {
		log.Fatal("Error to create table users", err)
	}

	_, err = db.Exec(`
create table if not exists expenses (
	id serial not null
		constraint expenses_pkey
			primary key,
	expense character varying,
    amount numeric DEFAULT '0.0' NOT NULL,
    created_at timestamp DEFAULT now() NOT NULL,
    user_id integer NOT NULL
);

`)
	if err != nil {
		log.Fatal("Error to create table expenses", err)
	}

	var countRow int
	err = db.QueryRow("SELECT COUNT(*) FROM accounts").Scan(&countRow)
	if err != nil {
		log.Fatal(err)
	}

	if countRow == 0 {
		query := `
INSERT INTO accounts
(slug, name, currency, balance, priority)
VALUES
('ukrbusd', 'Ukrsib B USD', 840, 100, 1),
('ukrbuah', 'Ukrsib B UAH', 980, 1000, 2),
('ukrpuah', 'Ukrsib P UAH', 980, 2000, 3),
('ukravto', 'Автомобіль', 980, -12000, 4)
;`
		_, err = db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}

	return db
}
