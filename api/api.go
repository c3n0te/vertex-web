package api

type User struct {
	Name  string `db:"name"`
	Email string `db:"email"`
}

type Satellite struct {
	NoradID uint32 `json:"noradid,omitempty" db:"noradid"` //NORAD Catalogue Number
	SatName string `json:"satname,omitempty" db:"satname"`
	Status  string `json:"status,omitempty" db:"status"`
	Line1   string `json:"line1,omitempty" db:"line1"`
	Line2   string `json:"line2,omitempty" db:"line2"`
}
