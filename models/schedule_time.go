package models

// ScheduleTime = specific times of day tied to a schedule
type ScheduleTime struct {
	ID         string `json:"id"`          // UUID
	ScheduleID string `json:"schedule_id"` // FK to schedules
	IntakeTime string `json:"intake_time"` // "HH:MM" format
}
