package models

import "time"

type Schedule struct {
	ID          string     `json:"id"`          // UUID
	MedicineID  string     `json:"medicine_id"` // FK to medicines
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Frequency   string     `json:"frequency"` // daily | weekly | custom
	TimesPerDay int        `json:"times_per_day"`
}
