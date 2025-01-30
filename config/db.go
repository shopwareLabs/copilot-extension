package config

import "github.com/philippgille/chromem-go"

func GetCollection(cfg *Info) (*chromem.Collection, error) {
	db, err := chromem.NewPersistentDB("./db", true)

	if err != nil {
		return nil, err
	}

	collection, err := db.GetOrCreateCollection("shopware_1", nil, chromem.NewEmbeddingFuncOllama("mxbai-embed-large", cfg.OllamaHost))
	if err != nil {
		return nil, err
	}

	return collection, nil
}
