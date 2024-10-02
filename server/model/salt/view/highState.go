package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
)

type HighState struct {
	Fun       string      `json:"fun" gorm:"->;type:varchar(50);not null;" example:"state.highstate"`
	JID       string      `json:"jid" gorm:"->;column:jid;type:varchar(20);not null;" example:"20060102150405999999"`
	Return    custom.JSON `json:"return" gorm:"->;type:jsonb;not null"`
	FullRet   custom.JSON `json:"full_ret" gorm:"->;type:jsonb"`
	ID        string      `json:"id" gorm:"->;primaryKey;type:varchar(255);not null;" example:"server.example.com"`
	Success   bool        `json:"success" gorm:"->;column:success;type:boolean;not null;" example:"true"`
	AlterTime *time.Time  `json:"alter_time" gorm:"->;type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();" example:"2006-01-02T15:04:05.999999-07:00"`
}

func (HighState) TableName() string {
	return "vw_salt_highstates"
}
