package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"vertex/api"
)

func ParseTLEFile() ([]api.Satellite, error) {
	tle_f, err := os.Open("./data/tle.txt")
	if err != nil {
		return nil, err
	}

	defer tle_f.Close()
	scanner := bufio.NewScanner(tle_f)
	sats := []api.Satellite{}
	sat := api.Satellite{}

	for scanner.Scan() {
		sat.SatName = scanner.Text()
		scanner.Scan()
		sat.Line1 = scanner.Text()
		scanner.Scan()
		l2 := scanner.Text()
		sat.Line2 = l2
		noradid, _ := strconv.ParseUint(strings.Split(l2, " ")[1], 10, 32)
		sat.NoradID = uint32(noradid)
		sat.Status = "online"
		sats = append(sats, sat)
	}

	return sats, nil
}

func ParseStationFile() ([]api.Station, error) {
	stn_data, err := os.ReadFile("./data/stations.json")
	if err != nil {
		return nil, err
	}

	var stns []api.Station
	json.Unmarshal(stn_data, &stns)
	return stns, nil
}
