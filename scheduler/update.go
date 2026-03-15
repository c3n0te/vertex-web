package main

import (
	"log/slog"
	"vertex/api"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func UpdateTasksStatus(db *sqlx.DB, tasks []api.Task, jobs []api.Job) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	scheduledTaskIDs := map[uuid.UUID]uint8{}
	for _, job := range jobs {
		scheduledTaskIDs[job.TaskID] = 1
	}

	for _, task := range tasks {
		val, ok := scheduledTaskIDs[task.TaskID]
		if val == 1 && ok {
			_, err = tx.NamedExec(
				`Update Tasks
					SET status='scheduled'
				WHERE taskid = :taskid`,
				&task,
			)
		} else {
			_, err = tx.NamedExec(
				`Update Tasks
					SET status='pending'
				WHERE taskid = :taskid`,
				&task,
			)
		}

		if err != nil {
			slog.Error("Failed to update task: ", "error", err)
			return err
		}
	}

	return tx.Commit()
}
