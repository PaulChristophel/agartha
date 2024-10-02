package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDB initializes the test database connection
func SetupTestDB() (*gorm.DB, error) {
	dsn := "host=db user=agartha password=agartha dbname=agartha port=5432 sslmode=disable"
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
}

func TestAuthUser_Create(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	// Ensure the table is clean before testing
	db.Exec("DELETE FROM auth_user")

	user := AuthUser{
		Password:    "plain_password", // Use the plain password here
		LastLogin:   nil,
		IsSuperuser: false,
		Username:    "testuser",
		FirstName:   "Test",
		LastName:    "User",
		Email:       "testuser@example.com",
		IsStaff:     false,
		IsActive:    true,
		DateJoined:  time.Now(),
	}

	err = user.Create(db)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, user.ID)

	// Verify user was created
	var createdUser AuthUser
	db.First(&createdUser, user.ID)
	assert.Equal(t, user.Username, createdUser.Username)
	assert.Equal(t, user.Email, createdUser.Email)
}

func TestAuthUser_Delete(t *testing.T) {
	db, err := SetupTestDB()
	assert.NoError(t, err)

	// Ensure the table is clean before testing
	db.Exec("DELETE FROM auth_user WHERE username = 'testuser'")

	user := AuthUser{
		Password:    "plain_password", // Use the plain password here
		LastLogin:   nil,
		IsSuperuser: false,
		Username:    "testuser",
		FirstName:   "Test",
		LastName:    "User",
		Email:       "testuser@example.com",
		IsStaff:     false,
		IsActive:    true,
		DateJoined:  time.Now(),
	}

	err = user.Create(db)
	assert.NoError(t, err)
	assert.NotEqual(t, 0, user.ID)

	// Delete the user
	err = user.Delete(db, user.ID)
	assert.NoError(t, err)

	// Verify user was deleted
	var deletedUser AuthUser
	result := db.First(&deletedUser, user.ID)
	assert.Error(t, result.Error)
	assert.Equal(t, gorm.ErrRecordNotFound, result.Error)
}
