package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
	"vertex/api"

	"github.com/alexedwards/argon2id"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func InsertSats(db *sqlx.DB) error {
	sats, err := ParseTLEFile()
	if err != nil {
		slog.Error("Failed to parse TLE file: ", "error", err)
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	for _, sat := range sats {
		_, err := tx.NamedExec(
			`INSERT INTO Satellites
			(noradid, satname, status, line1, line2)
			VALUES
			(:noradid, :satname, :status, :line1, :line2)`,
			&sat,
		)

		if err != nil {
			slog.Error("Failed to insert satellite", "error", err)
			return err
		}
	}

	slog.Info("Satellites inserted into DB")
	return tx.Commit()
}

func InsertStns(db *sqlx.DB) error {
	stns, err := ParseStationFile()
	if err != nil {
		slog.Error("Failed to parse Station file: ", "error", err)
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	for _, stn := range stns {
		_, err := tx.NamedExec(
			`INSERT INTO Stations
			(stnid, stnname, latitude, longitude, altitude, minhorizon, status)
			VALUES
			(:stnid, :stnname, :latitude, :longitude, :altitude, :minhorizon, :status)`,
			&stn,
		)

		if err != nil {
			slog.Error("Failed to insert station: ", "error", err)
			return err
		}
	}

	slog.Info("Stations inserted into DB")
	return tx.Commit()
}

func InsertUser(db *sqlx.DB, form url.Values) error {
	uname := form["username"][0]
	email := form["email"][0]
	pass := form["password"][0]
	re_pass := form["re_password"][0]

	if pass != re_pass {
		return &api.InvalidPassword{}
	}

	hash_pass, err := argon2id.CreateHash(pass, argon2id.DefaultParams)
	if err != nil {
		slog.Error("Failed to hash password: ", "error", err)
		return err
	}

	user := api.User{
		UserID:   uuid.New(),
		Username: uname,
		Email:    email,
		Password: hash_pass,
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	_, err = tx.NamedExec(
		`INSERT INTO Users
		(userid, username, email, password)
		VALUES
		(:userid, :username, :email, :password)`,
		&user,
	)

	if err != nil {
		slog.Error("Failed to insert user: ", "error", err)
		return err
	}

	slog.Info("User inserted into db.")
	return tx.Commit()
}

func InsertTask(db *sqlx.DB, form url.Values) error {
	plan := form["plan"][0]
	satname := form["satname"][0]
	zone, _ := time.Now().Zone()
	layout := "2006-01-02 15:04:00 MST"
	notb_date_str := fmt.Sprintf("%v %v:00 %s", form["notbefore_date"][0], form["notbefore_time"][0], zone)
	dead_date_str := fmt.Sprintf("%v %v:00 %s", form["deadline_date"][0], form["deadline_time"][0], zone)

	notb_date, err := time.Parse(layout, notb_date_str)
	if err != nil {
		slog.Error("Failed to create notbefore date: ", "error", err)
		return err
	}

	dead_date, err := time.Parse(layout, dead_date_str)
	if err != nil {
		slog.Error("Failed to create deadline date: ", "error", err)
		return err
	}

	priority, err := strconv.ParseInt(form["priority"][0], 10, 8)
	if err != nil {
		slog.Error("Failed to parse priority int: ", "error", err)
		return err
	}

	task := api.Task{
		TaskID:    uuid.New(),
		Plan:      plan,
		SatName:   satname,
		NotBefore: notb_date.UTC().Format(time.RFC3339),
		Deadline:  dead_date.UTC().Format(time.RFC3339),
		Priority:  uint8(priority),
		Status:    "pending",
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	_, err = tx.NamedExec(
		`INSERT INTO Tasks
		(taskid, plan, satname, notbefore, deadline, priority, status)
		VALUES
		(:taskid, :plan, :satname, :notbefore, :deadline, :priority, :status)`,
		&task,
	)

	if err != nil {
		slog.Error("Failed to insert task: ", "error", err)
		return err
	}

	slog.Info("Task inserted into DB")
	return tx.Commit()
}
