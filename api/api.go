package api

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Name  string `db:"name"`
	Email string `db:"email"`
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
	NotBefore time.Time `json:"notbefore,omitzero" db:"notbefore"`
	Deadline  time.Time `json:"deadline,omitzero" db:"deadline"`
	Priority  int8      `json:"priority,omitempty" db:"priority"`
}
