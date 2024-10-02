package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
)

type JID struct {
	JID       string      `json:"jid" gorm:"column:jid;type:varchar(20);autoIncrement:false;index:idx_jids_jid;primaryKey;not null" example:"20060102150405999999"`
	Load      custom.JSON `json:"load" gorm:"type:jsonb;not null"`
	AlterTime *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_jids_updated;not null" example:"2006-01-02T15:04:05.999999-07:00"`
}

func (JID) TableName() string {
	return "jids"
}
