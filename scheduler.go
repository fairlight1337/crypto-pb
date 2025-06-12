package main

import (
	"log"
	"time"

	"github.com/pocketbase/pocketbase"
)

// StartScheduler starts the periodic scanner job
func StartScheduler(app *pocketbase.PocketBase) {
	// Start immediately
	go scanAllWallets(app)

	// Then run every 10 minutes
	ticker := time.NewTicker(10 * time.Minute)
	for {
		select {
		case <-ticker.C:
			scanAllWallets(app)
		}
	}
}

func scanAllWallets(app *pocketbase.PocketBase) {
	log.Println("Starting wallet scan...")

	height, err := getCurrentBlockHeight()
	if err != nil {
		log.Printf("Failed to fetch block height: %v", err)
		return
	}

	records, err := app.Dao().FindRecordsByFilter(
		"cryptowallets",
		"1=1", // filter: mandatory non-empty filter string
		"",    // sort
		1,     // page
		0,     // perPage
	)
	if err != nil {
		log.Printf("Failed to fetch wallets: %v", err)
		return
	}

	for _, record := range records {
		address := record.GetString("address")
		currency := record.GetString("currency")

		log.Printf("Scanning wallet: %s (%s)", address, currency)

		switch currency {
		case "BTC":
			ScanBTCAddress(app, record, height)
		default:
			log.Printf("Currency %s not yet implemented", currency)
		}

		// simple throttle to avoid API rate limits
		time.Sleep(2 * time.Second)
	}

	log.Println("Wallet scan complete.")
}
