package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
)

type SaltCache struct {
	Bank      string       `gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_cache_bank"`
	PSQLKey   string       `gorm:"primaryKey;autoIncrement:false;type:varchar(255);not null;index:idx_salt_cache_psql_key"`
	Data      pgtype.JSONB `gorm:"type:jsonb;index:idx_salt_cache_data,type:gin,fast_update:on"`
	ID        uuid.UUID    `gorm:"type:uuid;DEFAULT:gen_random_uuid()"`
	AlterTime *time.Time   `gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_cache_updated"`
}

func (SaltCache) TableName() string {
	return "salt_cache"
}
