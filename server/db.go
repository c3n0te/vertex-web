package main

import (
	"log/slog"
	"vertex/api"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Migrate(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
      	name TEXT,
      	email TEXT
    );

	CREATE TABLE IF NOT EXISTS Stations (
    	stnid           TEXT UNIQUE PRIMARY KEY,
    	stnname         TEXT NOT NULL,
    	latitude        FLOAT NOT NULL,
    	longitude       FLOAT NOT NULL,
    	altitude        FLOAT NOT NULL,
    	minhorizon      FLOAT NOT NULL,
     	status			TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS Satellites (
        noradid         INTEGER UNIQUE PRIMARY KEY,
        satname         TEXT UNIQUE NOT NULL,
        status			TEXT NOT NULL,
        line1           TEXT NOT NULL,
        line2           TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS Tasks (
        taskid          TEXT UNIQUE PRIMARY KEY,
        userid          TEXT,
        stnid           TEXT,
        noradid         INTEGER,
        plan			TEXT NOT NULL,
        notbefore       TEXT NOT NULL,
        deadline        TEXT NOT NULL,
        priority        INTEGER NOT NULL,
        FOREIGN KEY(stnid) REFERENCES Stations(stnid),
        FOREIGN KEY(noradid) REFERENCES Satellites(noradid)
    );

    CREATE TABLE IF NOT EXISTS Passes (
        passid          TEXT UNIQUE PRIMARY KEY,
        stnid           TEXT NOT NULL,
        stnname         TEXT NOT NULL,
        noradid         INTEGER NOT NULL,
        satname         TEXT NOT NULL,
        azimuth         FLOAT NOT NULL,
        elevation       FLOAT NOT NULL,
        aos             TEXT NOT NULL,
        los             TEXT NOT NULL,
        FOREIGN KEY(stnid) REFERENCES Stations(stnid),
        FOREIGN KEY(noradid) REFERENCES Satellites(noradid)
    );

    CREATE TABLE IF NOT EXISTS Jobs (
        jobid           TEXT UNIQUE PRIMARY KEY,
        planid          TEXT UNIQUE NOT NULL,
        stnid           TEXT NOT NULL,
        noradid         INTEGER NOT NULL,
        satname         TEXT,
        azimuth         FLOAT,
        elevation       FLOAT,
        aos             TEXT NOT NULL,
        los             TEXT NOT NULL,
        priority        INTEGER NOT NULL
    );
    `
	db.MustExec(schema)
}

func InsertSats(db *sqlx.DB) error {
	sats, err := ParseTLEFile()
	if err != nil {
		slog.Error("Failed to parse TLE file")
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
			slog.Error("Failed to insert satellite")
			return err
		}
	}

	slog.Info("Satellites inserted into DB")
	return nil
}

func QuerySats(db *sqlx.DB) ([]api.Satellite, error) {
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
		slog.Error("Failed to query Satellites table")
		return nil, err
	}

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
