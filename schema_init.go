package main

import (
	"encoding/json"
	"log"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
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
        },{
                "name": "user",
                "type": "relation",
                "required": true,
                "options": { "collectionId": "users", "maxSelect": 1 }
        }]`

	var fields schema.Schema
	if err := json.Unmarshal([]byte(schema_json), &fields); err != nil {
		return err
	}

	rule := "user = @request.auth.id"

	coll := &models.Collection{
		Name:       "cryptowallets",
		Type:       models.CollectionTypeBase,
		Schema:     fields,
		ListRule:   types.Pointer(rule),
		ViewRule:   types.Pointer(rule),
		CreateRule: types.Pointer(rule),
		UpdateRule: types.Pointer(rule),
		DeleteRule: types.Pointer(rule),
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

	rule := "wallet.user = @request.auth.id"

	coll := &models.Collection{
		Name:       "cryptotransactions",
		Type:       models.CollectionTypeBase,
		Schema:     fields,
		ListRule:   types.Pointer(rule),
		ViewRule:   types.Pointer(rule),
		CreateRule: types.Pointer(rule),
		UpdateRule: types.Pointer(rule),
		DeleteRule: types.Pointer(rule),
	}

	return app.Dao().SaveCollection(coll)
}
