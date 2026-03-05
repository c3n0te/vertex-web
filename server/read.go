package main

import (
	"log/slog"
	"vertex/api"

	"github.com/jmoiron/sqlx"
)

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
