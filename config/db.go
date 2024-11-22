package config

import "github.com/philippgille/chromem-go"

func GetCollection() (*chromem.Collection, error) {
	db, err := chromem.NewPersistentDB("./db", true)

	if err != nil {
		return nil, err
	}

	collection, err := db.GetOrCreateCollection("shopware_1", nil, nil)
	if err != nil {
		return nil, err
	}

	return collection, nil
}
