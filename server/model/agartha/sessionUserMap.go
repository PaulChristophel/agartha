package model

import (
	"time"
)

// SessionUserMap represents the mapping of session identifiers to user IDs
type SessionUserMap struct {
	SessionID string    `gorm:"primaryKey;type:varchar(255);not null"`
	UserID    uint      `gorm:"index;not null"`
	CreatedAt time.Time `gorm:"type:timestamp with time zone;not null;default:now()"`
	User      AuthUser  `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (SessionUserMap) TableName() string {
	return "session_user_map"
}
