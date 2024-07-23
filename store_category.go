package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/badger/v3"
)

func addCategory(category Category) error {
	var err error

	err = validateCategory(category)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Unix()
	id := fmt.Sprintf("category:%d", timestamp)
	category.ID = id
	regdttm := now.Format("20060102150405")
	category.RegDTTM = regdttm

	err = db.Update(func(txn *badger.Txn) error {
		value, _ := json.Marshal(category)
		return txn.Set([]byte(id), value)
	})
	if err != nil {
		return err
	}

	return bleveIndex.Index(id, category)
}

func deleteCategory(id string) error {
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

func updateCategory(id string, updatedCategory Category) error {
	var err error

	var existingCategory Category
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &existingCategory)
		})
	})
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		updatedCategory.RegDTTM = existingCategory.RegDTTM
		updatedCategory.ID = existingCategory.ID
		value, _ := json.Marshal(updatedCategory)
		err = txn.Set([]byte(id), value)
		if err != nil {
			return err
		}

		// Update Bleve index
		return bleveIndex.Index(id, updatedCategory)
	})

	return err
}

func getCategoryList() ([]Category, error) {
	var results []Category = []Category{}

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte("category:")); it.ValidForPrefix([]byte("category:")); it.Next() {
			item := it.Item()
			var category Category

			err := item.Value(func(v []byte) error {
				return json.Unmarshal(v, &category)
			})
			if err != nil {
				return err
			}

			results = append(results, category)
		}

		return nil
	})

	if err != nil {
		return []Category{}, err
	}

	return results, nil
}

// Was for account. maybe necessary not
// func getCategoryListMAP() (map[string]Category, error) {
// 	var results map[string]Category = map[string]Category{}

// 	err := db.View(func(txn *badger.Txn) error {
// 		opts := badger.DefaultIteratorOptions
// 		opts.PrefetchSize = 10
// 		it := txn.NewIterator(opts)
// 		defer it.Close()

// 		for it.Seek([]byte("category:")); it.ValidForPrefix([]byte("category:")); it.Next() {
// 			item := it.Item()
// 			var category Category

// 			err := item.Value(func(v []byte) error {
// 				return json.Unmarshal(v, &category)
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			results[category.ID] = category
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		return map[string]Category{}, err
// 	}

// 	return results, nil
// }
