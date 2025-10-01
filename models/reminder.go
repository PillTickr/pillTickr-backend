package models

import "time"

// Reminder = an actual reminder instance generated for a schedule
type Reminder struct {
	ID               string     `json:"id"`                // UUID
	ScheduleID       string     `json:"schedule_id"`       // FK to schedules
	ReminderDatetime time.Time  `json:"reminder_datetime"` // exact datetime to remind
	Status           string     `json:"status"`            // pending | taken | missed
	TakenAt          *time.Time `json:"taken_at,omitempty"`
}
