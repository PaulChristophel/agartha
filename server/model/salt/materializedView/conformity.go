package model

import (
	"time"
)

type Conformity struct {
	ID             string     `json:"id" gorm:"->;primaryKey;type:varchar(255);not null;" example:"server.example.com"`
	AlterTime      *time.Time `json:"alter_time" gorm:"->;type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();" example:"2006-01-02T15:04:05.999999-07:00"`
	Success        bool       `json:"success" gorm:"->;column:success;type:boolean;not null" example:"true"`
	TrueCount      int64      `json:"true_count" gorm:"->;column:true_count;type:bigint;not null" example:"302"`
	FalseCount     int64      `json:"false_count" gorm:"->;column:false_count;type:bigint;not null" example:"0"`
	ChangedCount   int64      `json:"changed_count" gorm:"->;column:changed_count;type:bigint;not null" example:"12"`
	UnchangedCount int64      `json:"unchanged_count" gorm:"->;column:unchanged_count;type:bigint;not null" example:"290"`
}

func (Conformity) TableName() string {
	return "mat_conformity"
}
