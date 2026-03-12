package main

import (
	"log/slog"
	"vertex/api"

	"github.com/jmoiron/sqlx"
)

func InsertJobs(db *sqlx.DB, jobs []api.Job) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	for _, job := range jobs {
		_, err := tx.NamedExec(
			`INSERT INTO Jobs
			(jobid, taskid, stnid, stnname, noradid, satname, azimuth, elevation, aos, los, priority)
			VALUES
			(:jobid, :taskid, :stnid, :stnname, :noradid, :satname, :azimuth, :elevation, :aos, :los, :priority)`,
			&job,
		)

		if err != nil {
			slog.Error("Failed to insert job: ", "error", err)
			return err
		}
	}

	return tx.Commit()
}
