package main

import (
	"log/slog"
	"net/url"
	"sort"
	"vertex/api"

	"github.com/alexedwards/argon2id"
	"github.com/jmoiron/sqlx"
)

func ReadTaskCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(
		&count,
		`SELECT
			COUNT(*)
        FROM Tasks`,
	)

	if err != nil {
		slog.Error("Failed to query Jobs table: ", "error", err)
		return 0, err
	}

	return count, nil
}

func ReadJobCount(db *sqlx.DB) (int, error) {
	var count int
	err := db.Get(
		&count,
		`SELECT
			COUNT(*)
        FROM Jobs`,
	)

	if err != nil {
		slog.Error("Failed to query Jobs table: ", "error", err)
		return 0, err
	}

	return count, nil
}

func ReadTasks(db *sqlx.DB, pageSize int, offset int) ([]api.Task, error) {
	rows, err := db.Queryx(
		`SELECT
			taskid,
			satname,
			notbefore,
			deadline,
			priority
        FROM Tasks
        WHERE status='pending'
        LIMIT ?
        OFFSET ?`,
		pageSize,
		offset,
	)

	if err != nil {
		slog.Error("Failed to query Jobs table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	tasks := []api.Task{}
	task := api.Task{}
	for rows.Next() {
		err = rows.StructScan(&task)
		if err != nil {
			slog.Error("Failed to marshal db rows into Task struct: ", "error", err)
			return nil, err
		}

		tasks = append(tasks, task)
	}

	sort.Slice(tasks, func(i int, j int) bool {
		return tasks[i].Deadline < tasks[j].Deadline
	})

	return tasks, nil
}

func ReadJobs(db *sqlx.DB, pageSize int, offset int) ([]api.Job, error) {
	rows, err := db.Queryx(
		`SELECT
			jobid,
			taskid,
			stnid,
			stnname,
			noradid,
			satname,
			azimuth,
			elevation,
			aos,
			los,
			priority
        FROM Jobs
        LIMIT ?
        OFFSET ?`,
		pageSize,
		offset,
	)

	if err != nil {
		slog.Error("Failed to query Jobs table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	jobs := []api.Job{}
	job := api.Job{}
	for rows.Next() {
		err = rows.StructScan(&job)
		if err != nil {
			slog.Error("Failed to marshal db rows into Job struct: ", "error", err)
			return nil, err
		}

		jobs = append(jobs, job)
	}

	sort.Slice(jobs, func(i int, j int) bool {
		return jobs[i].AOS < jobs[j].AOS
	})

	return jobs, nil
}

func ReadUser(db *sqlx.DB, form url.Values) (*api.User, error) {
	email := form["email"][0]
	pass := form["password"][0]

	rows, err := db.Queryx(
		`SELECT
			userid,
			username,
			email,
			password
        FROM Users
        WHERE email = ?
        LIMIT 1`,
		email,
	)

	if err != nil {
		slog.Error("Failed to query Users table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	user := api.User{}
	for rows.Next() {
		err = rows.StructScan(&user)
		if err != nil {
			slog.Error("Failed to marshal db rows into User struct: ", "error", err)
			return nil, err
		}
	}

	if user.Email == "" {
		return &user, nil
	}

	match, err := argon2id.ComparePasswordAndHash(pass, user.Password)
	if err != nil {
		slog.Error("Failed to match passwords: ", "error", err)
		return nil, err
	}

	if !match {
		err := &api.InvalidPassword{}
		slog.Error("Failed to match passwords: ", "error", err)
		return nil, err
	}

	return &user, nil
}

func ReadSats(db *sqlx.DB) ([]api.Satellite, error) {
	rows, err := db.Query(
		`SELECT
      		noradid,
        	satname,
         	status,
          	line1,
           	line2
        FROM Satellites`,
	)

	if err != nil {
		slog.Error("Failed to query Satellites table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	sat := api.Satellite{}
	sats := []api.Satellite{}

	for rows.Next() {
		rows.Scan(
			&sat.NoradID,
			&sat.SatName,
			&sat.Status,
			&sat.Line1,
			&sat.Line2,
		)
		sats = append(sats, sat)
	}

	return sats, nil
}

func ReadStns(db *sqlx.DB) ([]api.Station, error) {
	rows, err := db.Query(
		`SELECT
			stnid,
			stnname,
			latitude,
			longitude,
			altitude,
			minhorizon,
			status
        FROM Stations`,
	)

	if err != nil {
		slog.Error("Failed to query Stations table: ", "error", err)
		return nil, err
	}

	defer rows.Close()
	stn := api.Station{}
	stns := []api.Station{}

	for rows.Next() {
		rows.Scan(
			&stn.StnID,
			&stn.StnName,
			&stn.Latitude,
			&stn.Longitude,
			&stn.Altitude,
			&stn.MinHorizon,
			&stn.Status,
		)
		stns = append(stns, stn)
	}

	return stns, nil
}
