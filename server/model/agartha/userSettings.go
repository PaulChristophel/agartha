package model

import (
	"time"

	"github.com/jackc/pgtype"
)

type UserSettings struct {
	UserID          uint         `gorm:"primaryKey"`
	Token           string       `gorm:"type:varchar(40);not null"`
	Created         time.Time    `gorm:"type:timestamp with time zone;not null"`
	SaltPermissions string       `gorm:"type:text;not null"`
	Settings        pgtype.JSONB `gorm:"type:jsonb;not null"`
	User            AuthUser     `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}
