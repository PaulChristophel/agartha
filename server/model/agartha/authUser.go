package model

import (
	"time"
)

type AuthUser struct {
	ID          uint       `gorm:"primaryKey"`
	Password    string     `gorm:"type:varchar(128);not null"`
	LastLogin   *time.Time `gorm:"type:timestamp with time zone;default:now();"`
	IsSuperuser bool       `gorm:"not null"`
	Username    string     `gorm:"type:varchar(150);not null;unique"`
	FirstName   string     `gorm:"type:varchar(150);not null"`
	LastName    string     `gorm:"type:varchar(150);not null"`
	Email       string     `gorm:"type:varchar(254);not null"`
	IsStaff     bool       `gorm:"not null"`
	IsActive    bool       `gorm:"not null"`
	DateJoined  time.Time  `gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (AuthUser) TableName() string {
	return "auth_user"
}
