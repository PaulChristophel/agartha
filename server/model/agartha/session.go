package model

import (
	"time"
)

// Session represents the structure of the sessions table in the database
type Session struct {
	ID        string    `gorm:"primaryKey;type:text;not null"`
	Data      string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone"`
	UpdatedAt time.Time `gorm:"type:timestamp with time zone"`
	ExpiresAt time.Time `gorm:"type:timestamp with time zone"`
}

func (Session) TableName() string {
	return "sessions"
}
