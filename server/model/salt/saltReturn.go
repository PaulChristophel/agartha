package model

import (
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
	"gorm.io/gorm"
)

type SaltReturn struct {
	Fun        string      `json:"fun" gorm:"type:varchar(50);not null;index:idx_salt_returns_fun" example:"event.fire"`
	JID        string      `json:"jid" gorm:"column:jid;type:varchar(20);not null;index:idx_salt_returns_jid;primaryKey" example:"20060102150405999999"`
	Return     custom.JSON `json:"return" gorm:"type:text;not null"`
	FullRet    custom.JSON `json:"full_ret" gorm:"type:text;not null"`
	ID         string      `json:"id" gorm:"type:varchar(255);not null;index:idx_salt_returns_id;primaryKey" example:"server.example.com"`
	Success    bool        `json:"success" gorm:"-" example:"true"`
	SuccessStr string      `json:"-" gorm:"column:success;type:varchar(10);not null;index:idx_salt_returns_success"`
	AlterTime  *time.Time  `json:"alter_time" gorm:"type:TIMESTAMP WITH TIME ZONE;DEFAULT:now();index:idx_salt_returns_updated" example:"2006-01-02T15:04:05.999999-07:00"`
}

// BeforeSave is a GORM hook that is called before saving a record to the database
func (sr *SaltReturn) BeforeSave(tx *gorm.DB) (err error) {
	if sr.Success {
		sr.SuccessStr = "true"
	} else {
		sr.SuccessStr = "false"
	}
	return nil
}

// AfterFind is a GORM hook that is called after retrieving a record from the database
func (sr *SaltReturn) AfterFind(tx *gorm.DB) (err error) {
	sr.Success = sr.SuccessStr == "true"
	return nil
}

func (SaltReturn) TableName() string {
	return "salt_returns"
}
