package model

import (
	"time"
)

// SessionUserMap represents the mapping of session identifiers to user IDs
type SessionUserMap struct {
	SessionID string    `json:"id" gorm:"primaryKey;type:varchar(255);not null"`
	UserID    uint      `json:"user_id" gorm:"index;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"type:timestamp with time zone;not null;default:now()"`
	User      AuthUser  `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (SessionUserMap) TableName() string {
	return "session_user_map"
}
