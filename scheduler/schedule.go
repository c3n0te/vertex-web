package main

import (
	"log/slog"
	"math/rand"
	"sort"
	"time"
	"vertex/api"

	"github.com/google/uuid"
)

func MakeConstraintsAndSets(passes []api.Pass) ([][]uint8, [][]uint8, map[uuid.UUID]struct{}, map[string]struct{}, time.Time, time.Time, error) {
	minAos := time.Now().UTC().AddDate(100, 0, 0)
	maxLos := time.Now().UTC().AddDate(-100, 0, 0)
	stnSet := map[uuid.UUID]struct{}{}
	satSet := map[string]struct{}{}

	for _, pass := range passes {
		aos, err := time.Parse(time.RFC3339, pass.AOS)
		if err != nil {
			slog.Error("Failed to create AOS date: ", "error", err)
			return nil, nil, nil, nil, time.Now(), time.Now(), err
		}

		aosUtc := aos.UTC().Round(time.Minute)
		if aosUtc.Before(minAos) {
			minAos = aosUtc
		}

		los, err := time.Parse(time.RFC3339, pass.LOS)
		if err != nil {
			slog.Error("Failed to create LOS date: ", "error", err)
			return nil, nil, nil, nil, time.Now(), time.Now(), err
		}

		losUtc := los.UTC().Round(time.Minute)
		if losUtc.After(maxLos) {
			maxLos = losUtc
		}

		stnSet[pass.StnID] = struct{}{}
		satSet[pass.SatName] = struct{}{}
	}

	dt := int(maxLos.Sub(minAos).Round(time.Minute).Minutes())
	numStns := len(stnSet)
	numSats := len(satSet)
	m1 := make([][]uint8, numStns)
	m2 := make([][]uint8, numSats)
	for i := range numStns {
		m1[i] = make([]uint8, dt)
	}

	for i := range numSats {
		m2[i] = make([]uint8, dt)
	}

	return m1, m2, stnSet, satSet, minAos, maxLos, nil
}

func MakeMaps(stnSet map[uuid.UUID]struct{}, satSet map[string]struct{}) (map[uuid.UUID]uint64, map[string]uint64) {
	stnMap := map[uuid.UUID]uint64{}
	satMap := map[string]uint64{}
	i := uint64(0)
	for stn, _ := range stnSet {
		stnMap[stn] = i
		i++
	}

	i = uint64(0)
	for sat, _ := range satSet {
		satMap[sat] = i
		i++
	}

	return stnMap, satMap
}

func MakeGraph(tasks []api.Task, passes []api.Pass) (map[api.Task][]api.Pass, error) {
	numTasks := len(tasks)
	numPasses := len(passes)
	G := map[api.Task][]api.Pass{}
	for _, task := range tasks {
		G[task] = []api.Pass{}
	}

	for i := range numTasks {
		notbDate, err := time.Parse(time.RFC3339, tasks[i].NotBefore)
		if err != nil {
			slog.Error("Failed to parse notbefore time: ", "error", err)
			return nil, err
		}

		deadDate, err := time.Parse(time.RFC3339, tasks[i].Deadline)
		if err != nil {
			slog.Error("Failed to parse deadline time: ", "error", err)
			return nil, err
		}

		for j := range numPasses {
			aos, err := time.Parse(time.RFC3339, passes[j].AOS)
			if err != nil {
				slog.Error("Failed to parse aos time: ", "error", err)
				return nil, err
			}

			los, err := time.Parse(time.RFC3339, passes[j].LOS)
			if err != nil {
				slog.Error("Failed to parse los time: ", "error", err)
				return nil, err
			}

			if tasks[i].SatName == passes[j].SatName && aos.After(notbDate) && los.Before(deadDate) {
				G[tasks[i]] = append(G[tasks[i]], passes[j])
			}
		}
	}

	return G, nil
}

func CheckNonzero(row []uint8, t0 int, tf int) bool {
	for i := t0; i < tf; i++ {
		if row[i] >= 1 {
			return true
		}
	}

	return false
}

func SetOne(row []uint8, t0 int, tf int) {
	for i := t0; i < tf; i++ {
		row[i] = 1
	}
}

func Schedule(tasks []api.Task, passes []api.Pass, maxIter int) ([]api.Job, error) {
	sort.Slice(tasks, func(i int, j int) bool {
		if tasks[i].Priority == tasks[j].Priority {
			return tasks[i].Deadline < tasks[j].Deadline
		}

		return tasks[i].Priority > tasks[j].Priority
	})

	stncon, satcon, stnset, satset, minAos, _, err := MakeConstraintsAndSets(passes)
	if err != nil {
		slog.Error("Failed to build constraint matrices and sets", "error", err)
		return nil, err
	}

	stnmap, satmap := MakeMaps(stnset, satset)
	G, err := MakeGraph(tasks, passes)
	if err != nil {
		slog.Error("Failed to create combination graph: ", "error", err)
		return nil, err
	}

	maxSum := uint64(0)
	matchings := map[api.Task]api.Pass{}
	tmp := map[api.Task]api.Pass{}
	for range maxIter {
		for _, task := range tasks {
			passes = G[task]
			filteredPasses := []api.Pass{}
			for _, pass := range passes {
				aos, err := time.Parse(time.RFC3339, pass.AOS)
				if err != nil {
					slog.Error("Failed to parse aos time: ", "error", err)
					return nil, err
				}

				los, err := time.Parse(time.RFC3339, pass.LOS)
				if err != nil {
					slog.Error("Failed to parse los time: ", "error", err)
					return nil, err
				}

				t0 := int(aos.Sub(minAos).Round(time.Minute).Minutes())
				tf := int(los.Sub(minAos).Round(time.Minute).Minutes())
				stnIdx := stnmap[pass.StnID]
				satIdx := satmap[pass.SatName]
				stnconRow := stncon[stnIdx]
				satconRow := satcon[satIdx]
				if CheckNonzero(stnconRow, t0, tf) {
					continue
				}

				if CheckNonzero(satconRow, t0, tf) {
					continue
				}

				filteredPasses = append(filteredPasses, pass)
			}

			numPasses := len(filteredPasses)
			if numPasses > 0 {
				pass := filteredPasses[rand.Intn(numPasses)]
				tmp[task] = pass
				aos, err := time.Parse(time.RFC3339, pass.AOS)
				if err != nil {
					slog.Error("Failed to parse aos time: ", "error", err)
					return nil, err
				}

				los, err := time.Parse(time.RFC3339, pass.LOS)
				if err != nil {
					slog.Error("Failed to parse los time: ", "error", err)
					return nil, err
				}

				t0 := int(aos.Sub(minAos).Round(time.Minute).Minutes())
				tf := int(los.Sub(minAos).Round(time.Minute).Minutes())
				stnIdx := stnmap[pass.StnID]
				satIdx := satmap[pass.SatName]
				stnconRow := stncon[stnIdx]
				satconRow := satcon[satIdx]
				SetOne(stnconRow, t0, tf)
				SetOne(satconRow, t0, tf)
			}
		}

		sum := uint64(0)
		for task, pass := range tmp {
			if pass.SatName != "" {
				sum += uint64(task.Priority)
			}
		}

		if sum > maxSum {
			maxSum = sum
			matchings = tmp
		}
	}

	jobs := []api.Job{}
	for task, pass := range matchings {
		job := api.Job{
			JobID:     uuid.New(),
			TaskID:    task.TaskID,
			StnID:     pass.StnID,
			StnName:   pass.StnName,
			NoradID:   pass.NoradID,
			SatName:   pass.SatName,
			Azimuth:   pass.Azimuth,
			Elevation: pass.Elevation,
			AOS:       pass.AOS,
			LOS:       pass.LOS,
			Priority:  task.Priority,
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}
