package models

import "time"

type Reminder struct {
	// ID                string     `json:"id"`
	UserID            string         `json:"user_id"`
	Name              string         `json:"name"`
	Notes             *string        `json:"notes,omitempty"`
	IsRecurring       bool           `json:"is_recurring"`
	RecurrencePattern *string        `json:"recurrence_pattern,omitempty"`
	StartDate         time.Time      `json:"start_date"`
	EndDate           *time.Time     `json:"end_date,omitempty"`
	IsActive          bool           `json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	Doses             []ReminderDose `json:"doses,omitempty"`
}

type ReminderDose struct {
	// ID         string    `json:"id"`
	ReminderID string    `json:"reminder_id"`
	Time       string    `json:"time"` // store in "15:04" format
	Dosage     string    `json:"dosage"`
	Notes      *string   `json:"notes,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
