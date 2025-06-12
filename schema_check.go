package main

import (
	"fmt"

	"github.com/pocketbase/pocketbase"
)

// Very basic sketch
func ValidateSchema(app *pocketbase.PocketBase) error {
	collection, err := app.Dao().FindCollectionByNameOrId("cryptowallets")
	if err != nil {
		return fmt.Errorf("collection cryptowallets missing: %w", err)
	}

	// Validate fields exist and types match
	fields := collection.Schema.AsMap()

	if f, ok := fields["address"]; !ok || f.Type != "text" {
		return fmt.Errorf("address field is missing or wrong type")
	}

	// Repeat for other fields...
	return nil
}
