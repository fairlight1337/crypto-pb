package main

import (
	"encoding/json"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func EnsureCollections(app *pocketbase.PocketBase) error {
	if err := ensureCryptowallets(app); err != nil {
		return err
	}
	if err := ensureCryptotransactions(app); err != nil {
		return err
	}
	return nil
}

func ensureCryptowallets(app *pocketbase.PocketBase) error {
	if _, err := app.Dao().FindCollectionByNameOrId("cryptowallets"); err == nil {
		log.Println("Collection 'cryptowallets' already exists")
		return nil
	}

	log.Println("Creating collection: cryptowallets")

	schema_json := `[{
		"name": "address",
		"type": "text",
		"required": true
	},{
		"name": "currency",
		"type": "select",
		"required": true,
		"options": { "maxSelect": 1, "values": ["BTC", "ETH"] }
	},{
		"name": "label",
		"type": "text"
	}]`

	var fields schema.Schema
	if err := json.Unmarshal([]byte(schema_json), &fields); err != nil {
		return err
	}

	coll := &models.Collection{
		Name:   "cryptowallets",
		Type:   models.CollectionTypeBase,
		Schema: fields,
	}

	return app.Dao().SaveCollection(coll)
}

func ensureCryptotransactions(app *pocketbase.PocketBase) error {
	if _, err := app.Dao().FindCollectionByNameOrId("cryptotransactions"); err == nil {
		log.Println("Collection 'cryptotransactions' already exists")
		return nil
	}

	log.Println("Creating collection: cryptotransactions")

	schema_json := `[{
		"name": "wallet",
		"type": "relation",
		"required": true,
		"options": { "collectionId": "cryptowallets" }
	},{
		"name": "txid",
		"type": "text",
		"required": true
	},{
		"name": "amount",
		"type": "number"
	},{
		"name": "timestamp",
		"type": "date"
	},{
		"name": "direction",
		"type": "select",
		"options": { "maxSelect": 1, "values": ["incoming", "outgoing"] }
	},{
		"name": "confirmed",
		"type": "bool"
	}]`

	var fields schema.Schema
	if err := json.Unmarshal([]byte(schema_json), &fields); err != nil {
		return err
	}

	coll := &models.Collection{
		Name:   "cryptotransactions",
		Type:   models.CollectionTypeBase,
		Schema: fields,
	}

	return app.Dao().SaveCollection(coll)
}
