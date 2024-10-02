package model

import (
	"time"

	"gorm.io/gorm"
)

type AuthUser struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Password    string     `json:"password" gorm:"type:varchar(128);not null"`
	LastLogin   *time.Time `json:"last_login" gorm:"type:timestamp with time zone;default:now();"`
	IsSuperuser bool       `json:"is_superuser" gorm:"not null"`
	Username    string     `json:"username" gorm:"type:varchar(150);not null;unique"`
	FirstName   string     `json:"first_name" gorm:"type:varchar(150);not null"`
	LastName    string     `json:"last_name" gorm:"type:varchar(150);not null"`
	Email       string     `json:"email" gorm:"type:varchar(254);not null"`
	IsStaff     bool       `json:"is_staff" gorm:"not null"`
	IsActive    bool       `json:"is_active" gorm:"not null"`
	DateJoined  time.Time  `json:"date_joined" gorm:"type:timestamp with time zone;not null;default:now();"`
}

func (AuthUser) TableName() string {
	return "auth_user"
}

func (user *AuthUser) Create(db *gorm.DB) error {
	// Use raw SQL to insert the user and hash the password using crypt and gen_salt because gorm doesn't support crypt directly
	result := db.Raw(`
		INSERT INTO auth_user (password, last_login, is_superuser, username, first_name, last_name, email, is_staff, is_active, date_joined)
		VALUES (crypt(?, gen_salt('bf', 8)), ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`, user.Password, user.LastLogin, user.IsSuperuser, user.Username, user.FirstName, user.LastName, user.Email, user.IsStaff, user.IsActive, user.DateJoined).Scan(&user.ID)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (user *AuthUser) Delete(db *gorm.DB, id uint) error {
	if err := db.Delete(&AuthUser{}, id).Error; err != nil {
		return err
	}

	return nil
}
