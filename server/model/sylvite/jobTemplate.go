package model

// JobTemplate represents the salt_job_template table
type JobTemplate struct {
	ID     int      `json:"id" gorm:"primaryKey;autoIncrement:true"`
	Name   string   `json:"name" gorm:"type:varchar(255);not null;index"` // Indexed
	Job    string   `json:"job" gorm:"type:jsonb;not null"`
	UserID uint     `json:"user_id" gorm:"not null;index"`
	User   AuthUser `gorm:"foreignKey:UserID;references:ID"` // Indexed
	Shared bool     `json:"shared" gorm:"index"`             // Indexed
}

func (JobTemplate) TableName() string {
	return "job_templates"
}
