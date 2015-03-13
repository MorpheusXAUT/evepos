package main

import (
	"log"
	"os"
	"runtime"

	"github.com/morpheusxaut/evepos/database"
	"github.com/morpheusxaut/evepos/mail"
	"github.com/morpheusxaut/evepos/misc"
	"github.com/morpheusxaut/evepos/session"
	"github.com/morpheusxaut/evepos/web"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config, err := misc.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: [%v]", err)
		os.Exit(2)
	}

	misc.SetupLogger(config.DebugLevel)

	db, err := database.SetupDatabase(config)
	if err != nil {
		misc.Logger.Criticalf("Failed to set up database: [%v]", err)
		os.Exit(2)
	}

	err = db.Connect()
	if err != nil {
		misc.Logger.Criticalf("Failed to connect to database: [%v]", err)
		os.Exit(2)
	}

	mailer := mail.SetupMailController(config, db)

	sessionController, err := session.SetupSessionController(config, db, mailer)
	if err != nil {
		misc.Logger.Criticalf("Failed to set up session controller: [%v]", err)
		os.Exit(2)
	}

	templates := web.SetupTemplates(db)

	checksums, err := web.SetupAssetChecksums()
	if err != nil {
		misc.Logger.Criticalf("Failed to calculate asset checkums: [%v]", err)
		os.Exit(2)
	}

	controller := web.SetupController(config, db, sessionController, templates, checksums)

	sessionController.StartRefreshTimer()

	controller.HandleRequests()
}
