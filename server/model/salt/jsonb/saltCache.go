package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
	"github.com/google/uuid"
)

type SaltCache struct {
	Bank      string      `json:"bank" gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_cache_bank"  example:"minions/server.example.com"`
	PSQLKey   string      `json:"psql_key" gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_cache_psql_key" example:"data"`
	Data      custom.JSON `json:"data" gorm:"type:jsonb;index:idx_salt_cache_data,type:gin,fast_update:on"`
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;DEFAULT:gen_random_uuid()" example:"123e4567-e89b-12d3-a456-426614174000"`
	AlterTime *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_cache_updated" example:"2006-01-02T15:04:05.999999-07:00"`
}

func (SaltCache) TableName() string {
	return "salt_cache"
}
