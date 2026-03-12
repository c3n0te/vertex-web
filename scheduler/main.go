package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func Wait(calcPeriod time.Time) {
	now := time.Now().UTC()
	if now.Before(calcPeriod) {
		diff := calcPeriod.Sub(now).Round(time.Second).Seconds()
		slog.Info(fmt.Sprintf("Sleeping %v seconds", diff))
		time.Sleep(time.Duration(diff) * time.Second)
	}
}

func main() {
	db, err := sqlx.Connect("sqlite3", "./vertex.db")
	if err != nil {
		slog.Error("Failed to open SQLite Database: ", "error", err)
		return
	}

	defer db.Close()
	err = godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file: ", "error", err)
		return
	}

	periodStr := os.Getenv("SCHEDULER_PERIOD")
	if periodStr == "" {
		slog.Error("Error loading period env var.")
		return
	}

	period, err := strconv.ParseInt(periodStr, 10, 64)
	if err != nil {
		slog.Error("Error parsing period env var.")
		return
	}

	interval := time.Duration(period)
	for {
		slog.Info("Calculating sleep period")
		calcPeriod := time.Now().UTC().Add(interval * time.Second)
		slog.Info("Reading Engine Notifications")
		notifs, err := ReadEngineNotifications(db)
		if err != nil {
			slog.Error("Failed to read notifications table", "error", err)
			Wait(calcPeriod)
			continue
		}

		if len(notifs) == 0 {
			slog.Info("No Engine Notifications")
			Wait(calcPeriod)
			continue
		}

		slog.Info("Reading Tasks")
		tasks, err := ReadTasks(db)
		if err != nil {
			slog.Error("Failed to read tasks table", "error", err)
			Wait(calcPeriod)
			continue
		}

		slog.Info("Reading Passes")
		passes, err := ReadPasses(db)
		if err != nil {
			slog.Error("Failed to read passes table", "error", err)
			Wait(calcPeriod)
			continue
		}

		slog.Info("Scheduling Jobs")
		jobs, err := Schedule(tasks, passes, 1000)
		if err != nil {
			slog.Error("Failed to schedule jobs", "error", err)
			Wait(calcPeriod)
			continue
		}

		slog.Info("Inserting Jobs")
		err = InsertJobs(db, jobs)
		if err != nil {
			slog.Error("Failed to insert jobs", "error", err)
			Wait(calcPeriod)
			continue
		}

		slog.Info("Deleting Engine Notifications")
		err = DeleteEngineNotifications(db)
		if err != nil {
			slog.Error("Failed to delete notifications", "error", err)
			Wait(calcPeriod)
			continue
		}

		slog.Info("Calculating remaining sleep time")
		Wait(calcPeriod)
	}
}
