package main

import (
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
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
