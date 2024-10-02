package model

import (
	"time"
)

// Session represents the structure of the sessions table in the database
type Session struct {
	ID        string    `json:"id" gorm:"primaryKey;type:text;not null"`
	Data      string    `json:"data" gorm:"type:text"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone"`
	UpdatedAt time.Time `json:"updated_at" gorm:"type:timestamp with time zone"`
	ExpiresAt time.Time `json:"expires_at" gorm:"type:timestamp with time zone"`
}

func (Session) TableName() string {
	return "sessions"
}
