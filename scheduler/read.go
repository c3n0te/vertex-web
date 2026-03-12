package main

import (
	"log/slog"
	"vertex/api"

	"github.com/jmoiron/sqlx"
)

func ReadTasks(db *sqlx.DB) ([]api.Task, error) {
	rows, err := db.Queryx(
		`SELECT
			taskid,
			plan,
			satname,
			notbefore,
			deadline,
			priority
        FROM Tasks`,
	)

	if err != nil {
		slog.Error("Failed to query Tasks table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	task := api.Task{}
	tasks := []api.Task{}
	for rows.Next() {
		err = rows.StructScan(&task)
		if err != nil {
			slog.Error("Failed to marshal db rows into Task struct: ", "error", err)
			return nil, err
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func ReadPasses(db *sqlx.DB) ([]api.Pass, error) {
	rows, err := db.Queryx(
		`SELECT
			passid,
			stnid,
			stnname,
			noradid,
			satname,
			azimuth,
			elevation,
			aos,
			los
        FROM Passes`,
	)

	if err != nil {
		slog.Error("Failed to query Passes table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	pass := api.Pass{}
	passes := []api.Pass{}
	for rows.Next() {
		err = rows.StructScan(&pass)
		if err != nil {
			slog.Error("Failed to marshal db rows into Pass struct: ", "error", err)
			return nil, err
		}

		passes = append(passes, pass)
	}

	return passes, nil
}

func ReadEngineNotifications(db *sqlx.DB) ([]api.Notification, error) {
	rows, err := db.Queryx(
		`SELECT
			service
        FROM Notifications
        WHERE service='engine'`,
	)

	if err != nil {
		slog.Error("Failed to query Notifications table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	notification := api.Notification{}
	notifications := []api.Notification{}
	for rows.Next() {
		err = rows.StructScan(&notification)
		if err != nil {
			slog.Error("Failed to marshal db rows into Notification struct: ", "error", err)
			return nil, err
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}
