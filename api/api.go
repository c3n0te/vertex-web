package api

import (
	"github.com/google/uuid"
)

type User struct {
	UserID   uuid.UUID `db:"userid"`
	Username string    `db:"username"`
	Email    string    `db:"email"`
	Password string    `db:"password"`
}

type Satellite struct {
	NoradID uint32 `json:"noradid,omitempty" db:"noradid"`
	SatName string `json:"satname,omitempty" db:"satname"`
	Status  string `json:"status,omitempty" db:"status"`
	Line1   string `json:"line1,omitempty" db:"line1"`
	Line2   string `json:"line2,omitempty" db:"line2"`
}

type Station struct {
	StnID      uuid.UUID `json:"stnid,omitempty" db:"stnid"`
	StnName    string    `json:"stnname,omitempty" db:"stnname"`
	Latitude   float32   `json:"latitude,omitempty" db:"latitude"`
	Longitude  float32   `json:"longitude,omitempty" db:"longitude"`
	Altitude   float32   `json:"altitude,omitempty" db:"altitude"`
	MinHorizon float32   `json:"minhorizon,omitempty" db:"minhorizon"`
	Status     string    `json:"status,omitempty" db:"status"`
}

type Task struct {
	TaskID    uuid.UUID `json:"taskid,omitempty" db:"taskid"`
	Plan      string    `json:"plan,omitempty" db:"plan"`
	SatName   string    `json:"satname,omitempty" db:"satname"`
	NotBefore string    `json:"notbefore,omitzero" db:"notbefore"`
	Deadline  string    `json:"deadline,omitzero" db:"deadline"`
	Priority  uint8     `json:"priority,omitempty" db:"priority"`
	Status    string    `json:"status" db:"status"`
}

type Pass struct {
	PassID    uuid.UUID `json:"passid,omitempty" db:"passid"`
	StnID     uuid.UUID `json:"stnid,omitempty" db:"stnid"`
	StnName   string    `json:"stnname,omitempty" db:"stnname"`
	NoradID   uint32    `json:"noradid,omitempty" db:"noradid"`
	SatName   string    `json:"satname,omitempty" db:"satname"`
	Azimuth   float32   `json:"azimuth,omitempty" db:"azimuth"`
	Elevation float32   `json:"elevation,omitempty" db:"elevation"`
	AOS       string    `json:"aos,omitzero" db:"aos"`
	LOS       string    `json:"los,omitzero" db:"los"`
}

type Job struct {
	JobID     uuid.UUID `json:"jobid,omitempty" db:"jobid"`
	TaskID    uuid.UUID `json:"taskid,omitempty" db:"taskid"`
	StnID     uuid.UUID `json:"stnid,omitempty" db:"stnid"`
	StnName   string    `json:"stnname,omitempty" db:"stnname"`
	NoradID   uint32    `json:"noradid,omitempty" db:"noradid"`
	SatName   string    `json:"satname,omitempty" db:"satname"`
	Azimuth   float32   `json:"azimuth,omitempty" db:"azimuth"`
	Elevation float32   `json:"elevation,omitempty" db:"elevation"`
	AOS       string    `json:"aos,omitzero" db:"aos"`
	LOS       string    `json:"los,omitzero" db:"los"`
	Priority  uint8     `json:"priority,omitempty" db:"priority"`
}

type Notification struct {
	Service string `json:"service,omitempty" db:"service"`
}
