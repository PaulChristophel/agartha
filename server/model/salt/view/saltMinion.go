package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
	"github.com/google/uuid"
)

type SaltMinion struct {
	MinionID  string      `json:"minion_id" gorm:"primaryKey;type:text;not null;" example:"server.example.com"`
	Grains    custom.JSON `json:"grains" gorm:"type:jsonb;"`
	Pillar    custom.JSON `json:"pillar" gorm:"type:jsonb;"`
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;" example:"123e4567-e89b-12d3-a456-426614174000"`
	AlterTime *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;" example:"2006-01-02T15:04:05.999999-07:00"`
}

func (SaltMinion) TableName() string {
	return "vw_salt_minions"
}
