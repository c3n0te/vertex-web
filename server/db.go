package main

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func Migrate(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS Users (
		userid			TEXT UNIQUE PRIMARY KEY,
      	username 		TEXT UNIQUE,
      	email 	 		TEXT UNIQUE,
        password  		TEXT
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
        satname			TEXT NOT NULL,
        plan			TEXT NOT NULL,
        notbefore       TEXT NOT NULL,
        deadline        TEXT NOT NULL,
        priority        INTEGER NOT NULL,
        Status			TEXT NOT NULL,
        FOREIGN KEY(satname) REFERENCES Satellites(satname)
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
        taskid          TEXT NOT NULL,
        stnid           TEXT NOT NULL,
        stnname			TEXT NOT NULL,
        noradid         INTEGER NOT NULL,
        satname         TEXT,
        azimuth         FLOAT,
        elevation       FLOAT,
        aos             TEXT NOT NULL,
        los             TEXT NOT NULL,
        priority        INTEGER NOT NULL
    );

    CREATE TABLE IF NOT EXISTS Parameters (
        max_horizon     INTEGER
    );

    CREATE TABLE IF NOT EXISTS Notifications (
        service         TEXT NOT NULL
    );

    CREATE TRIGGER IF NOT EXISTS parameter_trigger AFTER DELETE ON Parameters
    BEGIN
        INSERT INTO Parameters (max_horizon) VALUES (24);
    END;

    CREATE TRIGGER IF NOT EXISTS stn_insert_trigger AFTER INSERT ON Stations
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS stn_update_trigger AFTER UPDATE ON Stations
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS stn_delete_trigger AFTER DELETE ON Stations
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS sat_insert_trigger AFTER INSERT ON Satellites
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS sat_update_trigger AFTER UPDATE ON Satellites
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS sat_delete_trigger AFTER DELETE ON Satellites
    BEGIN
        INSERT INTO Notifications (service) VALUES ('orbital');
    END;

    CREATE TRIGGER IF NOT EXISTS pass_insert_trigger AFTER INSERT ON Passes
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    CREATE TRIGGER IF NOT EXISTS pass_update_trigger AFTER UPDATE ON Passes
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    CREATE TRIGGER IF NOT EXISTS pass_delete_trigger AFTER DELETE ON Passes
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    CREATE TRIGGER IF NOT EXISTS task_insert_trigger AFTER INSERT ON Tasks
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    CREATE TRIGGER IF NOT EXISTS task_update_trigger AFTER UPDATE ON Tasks
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    CREATE TRIGGER IF NOT EXISTS task_delete_trigger AFTER DELETE ON Tasks
    BEGIN
        INSERT INTO Notifications (service) VALUES ('engine');
    END;

    INSERT INTO Parameters (max_horizon) VALUES (24);
    `
	db.MustExec(schema)
}
