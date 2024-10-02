package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
)

type SaltEvent struct {
	Tag       string      `json:"tag" gorm:"type:varchar(255);not null;index:idx_salt_events_tag" example:"minion/refresh/server.example.com"`
	Data      custom.JSON `json:"data" gorm:"type:jsonb;not null"`
	AlterTime *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_events_alter_time" example:"2006-01-02T15:04:05.999999-07:00"`
	MasterID  string      `json:"master_id" gorm:"type:varchar(255);not null;index:idx_salt_events_master_id" example:"salt-f7884566d-td4gn_master"`
	ID        int64       `json:"id" gorm:"type:BIGSERIAL;NOT NULL;UNIQUE;index:idx_salt_events_id" example:"15167725"`
}

func (SaltEvent) TableName() string {
	return "salt_events"
}
