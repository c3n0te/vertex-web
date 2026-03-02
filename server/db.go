package main

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

func Migrate(db *sqlx.DB) {
	schema := `
    CREATE TABLE IF NOT EXISTS users (
      name TEXT,
      email TEXT
  );`
	db.MustExec(schema)
}

func Insert(db *sqlx.DB) error {
	user := User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
	}

	_, err := db.NamedExec(
		"INSERT INTO users (name, email) VALUES (:name, :email)",
		&user,
	)

	if err != nil {
		slog.Error("Failed to insert user")
		return err
	}

	slog.Info("User inserted into DB")
	return nil
}

func Query(db *sqlx.DB) ([]User, error) {
	rows, err := db.Query(
		`SELECT
      name,
      email
    FROM users`,
	)

	if err != nil {
		slog.Error("Failed to query users table")
		return nil, err
	}

	user := User{}
	users := []User{}

	for rows.Next() {
		rows.Scan(
			&user.Name,
			&user.Email,
		)
		users = append(users, user)
	}

	return users, nil
}
