package main

import (
	"log/slog"

	"github.com/jmoiron/sqlx"
)

func DeleteEngineNotifications(db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.Exec(
		`DELETE
		FROM Notifications
		WHERE service='engine'`,
	)

	if err != nil {
		slog.Error("Failed to delete engine notifications: ", "error", err)
		return err
	}

	return tx.Commit()
}
