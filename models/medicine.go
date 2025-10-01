package models

import "time"

type Medicine struct {
	ID           string    `json:"id"`      // UUID
	UserID       string    `json:"user_id"` // FK to users
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	Dosage       *string   `json:"dosage,omitempty"`       // e.g. "1 pill"
	Instructions *string   `json:"instructions,omitempty"` // e.g. "after meals"
	CreatedAt    time.Time `json:"created_at"`
}
