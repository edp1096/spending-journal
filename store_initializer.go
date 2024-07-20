package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
)

func initBadgerDB(password string) error {
	var err error

	saltFile := "salt"
	salt, err := getSalt(saltFile)
	if err != nil {
		return fmt.Errorf("failed to get salt: %w", err)
	}

	key := generateKey(password, salt)

	opts := badger.DefaultOptions("./badger_data")
	// opts.EncryptionKey = []byte("0123456789abcdefghijklmn") // 16 or 24 or 32 byte
	opts.EncryptionKey = key
	opts.IndexCacheSize = 100 << 20          // 100 MB
	opts.ValueLogFileSize = 64 * 1024 * 1024 // 64MB
	opts.ValueLogMaxEntries = 1000000
	opts.Logger = nil

	db, err = badger.Open(opts)
	if err != nil {
		db = nil
	}
	return err
}

func initBleveIndex() error {
	indexPath := "record_index.bleve"
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		mapping := bleve.NewIndexMapping()
		var err error
		bleveIndex, err = bleve.New(indexPath, mapping)
		if err != nil {
			return err
		}
	} else {
		bleveIndex, err = bleve.Open(indexPath)
		if err != nil {
			return err
		}
	}

	if db == nil {
		bleveIndex = nil
		return errors.New("db is not set")
	}

	return db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("record:")); it.ValidForPrefix([]byte("record:")); it.Next() {
			item := it.Item()
			err := item.Value(func(v []byte) error {
				var record Record
				if err := json.Unmarshal(v, &record); err != nil {
					return err
				}
				return bleveIndex.Index(string(item.Key()), record)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
