package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
)

type SaltKey struct {
	Bank      string      `json:"bank" gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_keys_bank" example:"pki/master/keys"`
	PSQLKey   string      `json:"psql_key" gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_keys_psql_key" example:"server.example.com"`
	Data      custom.JSON `json:"data" gorm:"type:jsonb;not null;index:idx_salt_keys_data,type:gin,fast_update:on"`
	AlterTime *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_keys_updated" example:"2006-01-02T15:04:05.999999-07:00"`
}

func (SaltKey) TableName() string {
	return "salt_keys"
}
