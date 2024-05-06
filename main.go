package main

import (
	"embed"
	"log"
	"os"

	"github.com/PaulChristophel/agartha/server/config"
	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/routes"
	"github.com/akamensky/argparse"
	// "github.com/akamensky/argparse"
)

//go:embed web/dist/*
var f embed.FS

//	@title			Agartha API
//	@version		1.0
//	@description	This is the Agartha API Backend

//	@contact.name	API Support
//	@contact.url	http://oit.gatech.edu
//	@contact.email	pmartin@gatech.edu

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	// Create new parser object
	parser := argparse.NewParser("agartha", "A web frontend for Salt and API for the Salt database.")

	parser.NewCommand("migrate", "Run migrations.")
	parser.NewCommand("serve", "Creates a listener service that runs the server.")

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil || len(os.Args) <= 1 {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		log.Print(parser.Usage(err))
		return
	}
	switch os.Args[1] {
	case "migrate":
		config.Config()
		db.ConnectToDatabase()
		db.Migrate()
	case "serve", "run":
		config.Config()
		db.ConnectToDatabase()
		routes.Router(f)
	default:
		log.Print(parser.Usage(err))
	}
}
