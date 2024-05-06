package model

import (
	"time"
)

type Conformity struct {
	ID             string     `gorm:"primaryKey;type:varchar(255);not null;index:idx_salt_returns_id"`
	AlterTime      *time.Time `gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_returns_updated"`
	Success        string     `gorm:"type:varchar(10);not null"`
	User           string     `gorm:"type:text"`
	TrueCount      int64      `gorm:"column:true_count;type:bigint;not null"`
	FalseCount     int64      `gorm:"column:false_count;type:bigint;not null"`
	ChangedCount   int64      `gorm:"column:changed_count;type:bigint;not null"`
	UnchangedCount int64      `gorm:"column:unchanged_count;type:bigint;not null"`
}

func (Conformity) TableName() string {
	return "mat_conformity"
}
