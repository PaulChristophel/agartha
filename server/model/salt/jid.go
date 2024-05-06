package model

import (
	"time"

	"github.com/jackc/pgtype"
)

type JID struct {
	JID       string      `gorm:"column:jid;type:varchar(20);primaryKey;autoIncrement:false"`
	Load      pgtype.JSON `gorm:"type:text;not null"`
	AlterTime *time.Time  `gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_jids_updated"`
}

func (JID) TableName() string {
	return "jids"
}
