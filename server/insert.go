package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"
	"vertex/api"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func InsertSats(db *sqlx.DB) error {
	sats, err := ParseTLEFile()
	if err != nil {
		slog.Error("Failed to parse TLE file: ", "error", err)
		return err
	}

	for _, sat := range sats {
		_, err := db.NamedExec(
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
	return nil
}

func InsertStns(db *sqlx.DB) error {
	stns, err := ParseStationFile()
	if err != nil {
		slog.Error("Failed to parse Station file: ", "error", err)
		return err
	}

	for _, stn := range stns {
		_, err := db.NamedExec(
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
	return nil
}

func InsertTask(db *sqlx.DB, task url.Values) error {
	plan := task["plan"][0]
	satname := task["satname"][0]
	zone, _ := time.Now().Zone()
	layout := "2006-01-02 15:04:00 MST"
	notb_date_str := fmt.Sprintf("%v %v:00 %s", task["notbefore_date"][0], task["notbefore_time"][0], zone)
	dead_date_str := fmt.Sprintf("%v %v:00 %s", task["deadline_date"][0], task["deadline_time"][0], zone)
	notb_date, err := time.Parse(layout, notb_date_str)
	if err != nil {
		slog.Error("Failed to create notbefore date: ", "error", err)
	}

	dead_date, err := time.Parse(layout, dead_date_str)
	if err != nil {
		slog.Error("Failed to create deadline date: ", "error", err)
	}

	priority, err := strconv.ParseInt(task["priority"][0], 10, 8)
	if err != nil {
		slog.Error("Failed to parse priority int: ", "error", err)
	}

	new_task := api.Task{
		TaskID:    uuid.New(),
		Plan:      plan,
		SatName:   satname,
		NotBefore: notb_date,
		Deadline:  dead_date,
		Priority:  int8(priority),
	}

	_, err = db.NamedExec(
		`INSERT INTO Tasks
		(taskid, plan, satname, notbefore, deadline, priority)
		VALUES
		(:taskid, :plan, :satname, :notbefore, :deadline, :priority)`,
		&new_task,
	)

	if err != nil {
		slog.Error("Failed to insert task: ", "error", err)
		return err
	}

	slog.Info("Task inserted into DB")
	return nil
}
