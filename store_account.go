package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
)

func addAccount(account Account) error {
	var err error

	err = validateAccount(account)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Unix()
	id := fmt.Sprintf("account:%d", timestamp)
	account.ID = id
	regdttm := now.Format("20060102150405")
	account.RegDTTM = regdttm

	err = db.Update(func(txn *badger.Txn) error {
		value, _ := json.Marshal(account)
		return txn.Set([]byte(id), value)
	})
	if err != nil {
		return err
	}

	return bleveIndex.Index(id, account)
}

func deleteAccount(id string) error {
	var err error

	// Remove Badger record
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Remove Bleve index
	err = bleveIndex.Delete(id)
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}

	return nil
}

func updateAccount(id string, updatedAccount Account) error {
	var err error

	var existingAccount Account
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &existingAccount)
		})
	})
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		updatedAccount.RegDTTM = existingAccount.RegDTTM
		updatedAccount.ID = existingAccount.ID
		value, _ := json.Marshal(updatedAccount)
		err = txn.Set([]byte(id), value)
		if err != nil {
			return err
		}

		// Update Bleve index
		return bleveIndex.Index(id, updatedAccount)
	})

	return err
}

func getAccountList() ([]Account, error) {
	var results []Account = []Account{}

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("account:")); it.ValidForPrefix([]byte("account:")); it.Next() {
			item := it.Item()
			var account Account

			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &account)
			})
			if err != nil {
				return err
			}

			results = append(results, account)
		}

		return nil
	})

	if err != nil {
		return []Account{}, err
	}

	return results, nil
}

func getAccountListMAP() (map[string]Account, error) {
	var results map[string]Account = map[string]Account{}

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("account:")); it.ValidForPrefix([]byte("account:")); it.Next() {
			item := it.Item()
			var account Account

			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &account)
			})
			if err != nil {
				return err
			}

			// results = append(results, account)
			results[account.ID] = account
		}

		return nil
	})

	if err != nil {
		return map[string]Account{}, err
	}

	return results, nil
}
