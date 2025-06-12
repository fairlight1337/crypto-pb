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

// Network configuration is set at runtime
var (
	useTestnet         bool
	blockstreamBaseUrl string
)

func configureNetwork(testnet bool) {
	useTestnet = testnet
	if useTestnet {
		blockstreamBaseUrl = "https://blockstream.info/testnet/api"
	} else {
		blockstreamBaseUrl = "https://blockstream.info/api"
	}
}

// Transaction response from Blockstream API (partial)
type Tx struct {
	TxID   string `json:"txid"`
	Status struct {
		BlockHeight int   `json:"block_height"`
		BlockTime   int64 `json:"block_time"`
		Confirmed   bool  `json:"confirmed"`
	} `json:"status"`
	Vin []struct {
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
func ScanBTCAddress(app *pocketbase.PocketBase, wallet *models.Record, currentHeight int) {
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
		processTransaction(app, wallet, address, tx, currentHeight)
	}

	updateWalletBalance(app, wallet)
}

// Process single transaction & insert into PocketBase if not present
func processTransaction(app *pocketbase.PocketBase, wallet *models.Record, address string, tx Tx, currentHeight int) {
	txid := tx.TxID

	// Check if tx already exists so we can update confirmations
	existing, err := app.Dao().FindFirstRecordByFilter("cryptotransactions", "txid = {:txid}", map[string]any{"txid": txid})

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

	confirmations := 0
	if tx.Status.Confirmed {
		confirmations = currentHeight - tx.Status.BlockHeight + 1
		if confirmations < 0 {
			confirmations = 0
		}
	}
	timestamp := time.Unix(tx.Status.BlockTime, 0)

	if err == nil && existing != nil {
		updated := false
		if existing.GetInt("confirmations") != confirmations {
			existing.Set("confirmations", confirmations)
			updated = true
		}
		if existing.GetDateTime("timestamp").IsZero() && tx.Status.BlockTime != 0 {
			existing.Set("timestamp", timestamp)
			updated = true
		}
		if updated {
			if err := app.Dao().SaveRecord(existing); err != nil {
				log.Printf("Failed to update tx %s: %v", txid, err)
			}
		}
		return
	}

	collection, err := app.Dao().FindCollectionByNameOrId("cryptotransactions")
	if err != nil {
		log.Printf("Failed to load collection: %v", err)
		return
	}

	newTx := models.NewRecord(collection)
	newTx.Set("wallet", wallet.Id)
	newTx.Set("txid", txid)
	newTx.Set("amount", amount)
	newTx.Set("timestamp", timestamp)
	newTx.Set("direction", direction)
	newTx.Set("confirmations", confirmations)

	if err := app.Dao().SaveRecord(newTx); err != nil {
		log.Printf("Failed to insert tx %s: %v", txid, err)
	} else {
		log.Printf("Inserted new transaction: %s", txid)
	}
}

// updateWalletBalance fetches the current wallet balance and saves it
func updateWalletBalance(app *pocketbase.PocketBase, wallet *models.Record) {
	address := wallet.GetString("address")
	url := fmt.Sprintf("%s/address/%s", blockstreamBaseUrl, address)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Failed to fetch balance for %s: %v", address, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Unexpected response for balance: %s", body)
		return
	}

	var data struct {
		ChainStats struct {
			FundedTxoSum int64 `json:"funded_txo_sum"`
			SpentTxoSum  int64 `json:"spent_txo_sum"`
		} `json:"chain_stats"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Printf("Failed to decode balance for %s: %v", address, err)
		return
	}

	balance := float64(data.ChainStats.FundedTxoSum-data.ChainStats.SpentTxoSum) / 1e8
	if wallet.GetFloat("balance") == balance {
		return
	}
	wallet.Set("balance", balance)
	if err := app.Dao().SaveRecord(wallet); err != nil {
		log.Printf("Failed to update balance for %s: %v", address, err)
	}
}

// getCurrentBlockHeight returns the tip height for the configured network
func getCurrentBlockHeight() (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/blocks/tip/height", blockstreamBaseUrl))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected response: %s", body)
	}
	var height int
	if err := json.NewDecoder(resp.Body).Decode(&height); err != nil {
		return 0, err
	}
	return height, nil
}
