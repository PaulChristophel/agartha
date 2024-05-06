package model

import "time"

type SaltEvent struct {
	ID        int64      `gorm:"type:BIGINT;NOT NULL;UNIQUE;DEFAULT:nextval('seq_salt_events_id')"`
	Tag       string     `gorm:"type:varchar(255);not null;index:idx_salt_events_tag"`
	Data      string     `gorm:"type:text;not null"`
	AlterTime *time.Time `gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now()"`
	MasterID  string     `gorm:"type:varchar(255);not null"`
}

func (SaltEvent) TableName() string {
	return "salt_events"
}
