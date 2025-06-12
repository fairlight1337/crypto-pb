package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
)

// Global: switch between mainnet and testnet here
const useTestnet = true

var blockstreamBaseUrl string

func init() {
	if useTestnet {
		blockstreamBaseUrl = "https://blockstream.info/testnet/api"
	} else {
		blockstreamBaseUrl = "https://blockstream.info/api"
	}
}

// Transaction response from Blockstream API (partial)
type Tx struct {
	TxID   string `json:"txid"`
	Height int    `json:"status.height"`
	Time   int64  `json:"status.block_time"`
	Vin    []struct {
		Prevout struct {
			ScriptpubkeyAddress string `json:"scriptpubkey_address"`
			Value               int64  `json:"value"`
		} `json:"prevout"`
	} `json:"vin"`
	Vout []struct {
		ScriptpubkeyAddress string `json:"scriptpubkey_address"`
		Value               int64  `json:"value"`
	} `json:"vout"`
}

// Entry point for scanning a single BTC address
func ScanBTCAddress(app *pocketbase.PocketBase, wallet *models.Record) {
	address := wallet.GetString("address")

	url := fmt.Sprintf("%s/address/%s/txs", blockstreamBaseUrl, address)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch transactions: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Unexpected response: %s", body)
		return
	}

	var txs []Tx
	if err := json.NewDecoder(resp.Body).Decode(&txs); err != nil {
		log.Printf("Failed to decode transactions: %v", err)
		return
	}

	for _, tx := range txs {
		processTransaction(app, wallet, address, tx)
	}
}

// Process single transaction & insert into PocketBase if not present
func processTransaction(app *pocketbase.PocketBase, wallet *models.Record, address string, tx Tx) {
	txid := tx.TxID

	// Deduplication: check if already exists
	existing, err := app.Dao().FindFirstRecordByFilter("cryptotransactions", "txid = {:txid}", map[string]any{"txid": txid})
	if err == nil && existing != nil {
		return
	}

	direction := "incoming"
	foundAsOutput := false
	var amountSats int64

	for _, vout := range tx.Vout {
		if vout.ScriptpubkeyAddress == address {
			amountSats += vout.Value
			foundAsOutput = true
		}
	}
	if !foundAsOutput {
		direction = "outgoing"
		for _, vin := range tx.Vin {
			if vin.Prevout.ScriptpubkeyAddress == address {
				amountSats -= vin.Prevout.Value
			}
		}
	}

	amount := float64(amountSats) / 1e8

	collection, err := app.Dao().FindCollectionByNameOrId("cryptotransactions")
	if err != nil {
		log.Printf("Failed to load collection: %v", err)
		return
	}

	newTx := models.NewRecord(collection)
	newTx.Set("wallet", wallet.Id)
	newTx.Set("txid", txid)
	newTx.Set("amount", amount)
	newTx.Set("timestamp", time.Unix(tx.Time, 0))
	newTx.Set("direction", direction)
	newTx.Set("confirmed", tx.Height > 0)

	if err := app.Dao().SaveRecord(newTx); err != nil {
		log.Printf("Failed to insert tx %s: %v", txid, err)
	} else {
		log.Printf("Inserted new transaction: %s", txid)
	}
}
