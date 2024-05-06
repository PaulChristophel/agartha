package db

import (
	"fmt"
	"log"
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/spf13/viper"
)

// Declare the variable for the database
var DB *gorm.DB

func ConnectToDatabase() {
	var err error
	p := viper.GetString("DB_PORT")
	port, err := strconv.ParseUint(p, 10, 32)

	if err != nil {
		log.Println("Idiot")
	}

	// Connection URL to connect to Postgres Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", viper.GetString("DB_HOST"), port, viper.GetString("DB_USER"), viper.GetString("DB_PASSWORD"), viper.GetString("DB_NAME"), viper.GetString("DB_SSLMODE"))
	// Connect to the DB and initialize the DB variable
	DB, err = gorm.Open(postgres.Open(dsn))

	if err != nil {
		panic("failed to connect database")
	}

	log.Print("Connection Opened to Database")
}
