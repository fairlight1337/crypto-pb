package main

import (
	"log"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	askNetwork()

	app := pocketbase.New()

	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		// 1️⃣ Ensure collections exist
		if err := EnsureCollections(app); err != nil {
			log.Fatal("Failed to setup collections:", err)
		}

		// 2️⃣ (Optional) Validate schema integrity
		if err := ValidateSchema(app); err != nil {
			log.Fatal("Schema invalid:", err)
		}

		// 3️⃣ Start scheduler
		go StartScheduler(app)

		return nil
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}

func askNetwork() {
	choice := "testnet"
	prompt := &survey.Select{
		Message: "Select BTC network:",
		Options: []string{"testnet", "mainnet"},
		Default: "testnet",
	}
	_ = survey.AskOne(prompt, &choice)
	configureNetwork(choice == "testnet")
}
