package main

import (
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/routes"
	"github.com/akamensky/argparse"
)

//go:embed web/dist/*
var dist embed.FS

//	@title			Agartha API
//	@version		1.0
//	@description	This is the Agartha API Backend

//	@contact.name	API Support
//	@contact.url	https://github.com/PaulChristophel/agartha/issues
//	@contact.email	kind.frog8344@fastmail.com

//	@license.name	AGPL 3.0
//	@license.url	https://www.gnu.org/licenses/agpl-3.0.html#license-text

// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	// Create new parser object
	parser := argparse.NewParser("agartha", "A web frontend for Salt and API for the Salt database.")

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
		db.ConnectToDatabase(config.AgarthaConfig.DB)
		if config.AgarthaConfig.DB.Tables.UseJSONB {
			err = db.MigrateJSONB(config.AgarthaConfig.DB.Tables)
		} else {
			err = db.Migrate(config.AgarthaConfig.DB.Tables)
		}
		if err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migration completed successfully.")
	case serveCmd.Happened():
		db.ConnectToDatabase(config.AgarthaConfig.DB)
		err = routes.Router(dist, *config.AgarthaConfig)
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
