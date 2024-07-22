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

func changePassword(oldPassword, newPassword string) error {
	oldSalt, err := getSalt("salt")
	if err != nil {
		return fmt.Errorf("failed to get old salt: %w", err)
	}
	oldKey := generateKey(oldPassword, oldSalt)

	// oldOpts := badger.DefaultOptions("./badger_data").WithEncryptionKey(oldKey).WithIndexCacheSize(100 << 20)
	oldOpts := badger.DefaultOptions("./badger_data")
	oldOpts.EncryptionKey = oldKey
	oldOpts.IndexCacheSize = 100 << 20          // 100 MB
	oldOpts.ValueLogFileSize = 64 * 1024 * 1024 // 64MB
	oldOpts.ValueLogMaxEntries = 1000000
	oldOpts.Logger = nil
	oldDB, err := badger.Open(oldOpts)
	if err != nil {
		return fmt.Errorf("failed to open old DB: %w", err)
	}
	defer oldDB.Close()

	if err := os.Remove("salt"); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old salt file: %w", err)
	}

	newSalt, err := getSalt("salt")
	if err != nil {
		return fmt.Errorf("failed to generate new salt: %w", err)
	}

	newKey := generateKey(newPassword, newSalt)
	// newOpts := badger.DefaultOptions("./new_badger_data").WithEncryptionKey(newKey).WithIndexCacheSize(100 << 20)
	newOpts := badger.DefaultOptions("./new_badger_data")
	newOpts.EncryptionKey = newKey
	newOpts.IndexCacheSize = 100 << 20          // 100 MB
	newOpts.ValueLogFileSize = 64 * 1024 * 1024 // 64MB
	newOpts.ValueLogMaxEntries = 1000000
	newOpts.Logger = nil
	newDB, err := badger.Open(newOpts)
	if err != nil {
		return fmt.Errorf("failed to create new DB: %w", err)
	}
	defer newDB.Close()

	err = oldDB.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				return newDB.Update(func(txn *badger.Txn) error {
					return txn.Set(item.Key(), val)
				})
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	oldDB.Close()
	oldDB = nil
	newDB.Close()
	newDB = nil
	if err := os.RemoveAll("./badger_data"); err != nil {
		return fmt.Errorf("failed to remove old DB: %w", err)
	}
	if err := os.Rename("./new_badger_data", "./badger_data"); err != nil {
		return fmt.Errorf("failed to rename new DB: %w", err)
	}

	return nil
}
