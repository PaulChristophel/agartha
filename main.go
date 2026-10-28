package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/PaulChristophel/sylvite/server/config"
	"github.com/PaulChristophel/sylvite/server/db"
	"github.com/PaulChristophel/sylvite/server/routes"
	"github.com/akamensky/argparse"
)

//go:embed web/dist/*
var dist embed.FS

//	@title			Sylvite API
//	@version		1.0
//	@description	A Salt management console

//	@contact.name	API Support
//	@contact.url	https://github.com/PaulChristophel/sylvite/issues
//	@contact.email	kind.frog8344@fastmail.com

//	@license.name	AGPL 3.0
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.html#license-text

// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	// Create new parser object
	parser := argparse.NewParser("sylvite", "A Salt management console")

	// Define commands
	migrateCmd := parser.NewCommand("migrate", "Run migrations.")
	serveCmd := parser.NewCommand("serve", "Creates a listener service that runs the server.")
	versionCmd := parser.NewCommand("version", "Returns the version of the server.")

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil || len(os.Args) <= 1 {
		// In case of error or no arguments, print usage
		fmt.Print(parser.Usage(err))
		return
	}

	// Initialize config
	err = config.InitConfig()
	if err != nil {
		fmt.Print(err)
		return
	}

	// Handle commands
	switch {
	case migrateCmd.Happened():
		if err := config.SylviteConfig.ValidateDatabase(); err != nil {
			log.Fatalf("Invalid configuration: %v", err)
		}
		if err := db.ConnectToDatabase(config.SylviteConfig.DB); err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}
		if config.SylviteConfig.DB.Tables.UseJSONB {
			err = db.MigrateJSONB(config.SylviteConfig.DB.Tables)
		} else {
			err = db.Migrate(config.SylviteConfig.DB.Tables)
		}
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migration completed successfully.")
	case serveCmd.Happened():
		if err := config.SylviteConfig.ValidateForServe(); err != nil {
			log.Fatalf("Invalid configuration: %v", err)
		}
		if err := db.ConnectToDatabase(config.SylviteConfig.DB); err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}
		err = routes.Router(dist, *config.SylviteConfig)
		if err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	case versionCmd.Happened():
		routes.PrintVersion()
	default:
		log.Printf("Unknown command\n")
		log.Println(parser.Usage(nil))
		os.Exit(1)
	}
}
