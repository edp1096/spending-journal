package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
)

func addRecord(record Record) error {
	var err error

	err = validateRecord(record)
	if err != nil {
		return err
	}

	now := time.Now()
	timestamp := now.Unix()
	// key := fmt.Sprintf("record:%d", timestamp)
	id := fmt.Sprintf("record:%d", timestamp)
	record.ID = id

	regdttm := now.Format("20060102150405")
	record.RegDTTM = regdttm

	err = db.Update(func(txn *badger.Txn) error {
		value, _ := json.Marshal(record)
		// return txn.Set([]byte(key), value)
		return txn.Set([]byte(id), value)
	})
	if err != nil {
		return err
	}

	return bleveIndex.Index(id, record)
}

func searchRecordsByDescription(queries []string, page, pageSize int, queryType string) ([]Record, float64, float64, int, error) {
	var results []Record
	var totalPay float64 = 0
	var totalIncome float64 = 0

	boolQuery := bleve.NewBooleanQuery()
	for _, query := range queries {
		matchQuery := bleve.NewMatchQuery(query)

		if queryType == "AND" {
			boolQuery.AddMust(matchQuery)
		} else {
			boolQuery.AddShould(matchQuery)
		}
	}

	search := bleve.NewSearchRequest(boolQuery)
	search.Size = pageSize
	search.From = (page - 1) * pageSize
	search.SortBy([]string{"-_score"})

	searchResults, err := bleveIndex.Search(search)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	for _, hit := range searchResults.Hits {
		var record Record
		err := db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(hit.ID))
			if err != nil {
				return err
			}

			return item.Value(func(v []byte) error {
				return json.Unmarshal(v, &record)
			})
		})
		if err != nil {
			continue
		}

		results = append(results, record)

		switch record.TransactionType {
		case "record_type_pay":
			totalPay += record.Amount
		case "record_type_income":
			totalIncome += record.Amount
		}
	}

	return results, totalPay, totalIncome, int(searchResults.Total), nil
}

func getSumByPeriod(startDate, endDate string) ([]Record, float64, float64, error) {
	var records []Record
	var sumPay float64 = 0
	var sumIncome float64 = 0

	err := db.View(func(txn *badger.Txn) error {
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
				if record.Date >= startDate && record.Date <= endDate {
					records = append(records, record)

					switch record.TransactionType {
					case "record_type_pay":
						sumPay += record.Amount
					case "record_type_income":
						sumIncome += record.Amount
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return records, sumPay, sumIncome, err
}
