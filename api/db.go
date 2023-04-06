package api

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v4"
)

func (p *Person) MarshalBinary() ([]byte, error) {
	// Encode the struct as JSON
	jsonData, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	// Encode the length of the JSON data as a 4-byte little-endian integer
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, uint32(len(jsonData))); err != nil {
		return nil, err
	}

	// Concatenate the length and JSON data and return the result
	return append(buf.Bytes(), jsonData...), nil
}

func (p *Person) UnmarshalBinary(data []byte) error {
	// Read the first 4 bytes as a little-endian integer, which represents the length of the JSON data
	buf := bytes.NewReader(data)
	var length uint32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return err
	}

	// Decode the remaining data as JSON
	jsonData := make([]byte, length)
	if _, err := buf.Read(jsonData); err != nil {
		return err
	}

	// Unmarshal the JSON data into the struct
	if err := json.Unmarshal(jsonData, p); err != nil {
		return err
	}

	return nil
}

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
