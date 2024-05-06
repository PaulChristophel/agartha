package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

type SaltReturn struct {
	Fun       string      `gorm:"type:varchar(50);not null;index:idx_salt_returns_fun"`
	JID       string      `gorm:"column:jid;type:varchar(255);not null;index:idx_salt_returns_jid"`
	Return    pgtype.JSON `gorm:"type:text;not null"`
	FullRet   pgtype.JSON `gorm:"type:text"`
	ID        string      `gorm:"type:varchar(255);not null;index:idx_salt_returns_id"`
	Success   string      `gorm:"type:varchar(10);not null"`
	AlterTime *time.Time  `gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_returns_updated"`
	UUID      uuid.UUID   `gorm:"primaryKey;type:uuid;DEFAULT:gen_random_uuid()"`
}

func (SaltReturn) TableName() string {
	return "salt_returns"
}
