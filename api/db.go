package api

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v4"
	"log"
)

func LoadBadgerDB(config ApiConfig) ApiConfig {
	opts := badger.DefaultOptions(config.DataBaseFile)
	db, err := badger.Open(opts)
	CheckAndLog(err, "error opening badgerdb", config)
	defer db.Close()

	txn := db.NewTransaction(true)
	defer txn.Discard()

	if _, err := txn.Get([]byte("data")); err != nil && err == badger.ErrKeyNotFound {
		log.Printf("creating %s database", config.DataBaseFile)
		err = txn.Set([]byte("data"), []byte("{}"))
		CheckAndLog(err, "error creating badgerdb", config)
	}
	if err != nil {
		log.Println(err)
	}

	item, err := txn.Get([]byte("data"))
	if err != nil {
		log.Println(err)
	}

	err = item.Value(func(val []byte) error {
		return json.Unmarshal(val, &config.DataBase)
	})
	CheckAndLog(err, "error decoding badgerdb", config)

	log.Println("loading badgerdb database from file")
	return config
}

func DefaultSaveBadgerDB(config ApiConfig) {
  db, err := badger.Open(badger.DefaultOptions(config.DataBaseFile))
  CheckAndLog(err, "error opening badgerdb", config)
  defer db.Close()

  txn := db.NewTransaction(true)
  defer txn.Discard()

  jsonBytes, err := json.MarshalIndent(config.DataBase, "", "\t")
  CheckAndLog(err, "error saving the database to file", config)

  err = txn.Set([]byte("data"), jsonBytes)
  CheckAndLog(err, "error setting value in badgerdb", config)

  err = txn.Commit()
  CheckAndLog(err, "error committing transaction in badgerdb", config)

  log.Println("Saving badgerdb to file")
}
