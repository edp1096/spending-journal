package server

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	id := fmt.Sprintf("record:%d", timestamp)
	record.ID = id
	regdttm := now.Format("20060102150405")
	record.RegDTTM = regdttm

	err = db.Update(func(txn *badger.Txn) error {
		value, _ := json.Marshal(record)
		return txn.Set([]byte(id), value)
	})
	if err != nil {
		return err
	}

	return bleveIndex.Index(id, record)
}

func deleteRecord(id string) error {
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

func updateRecord(id string, updatedRecord Record) error {
	var err error

	var existingRecord Record
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(id))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &existingRecord)
		})
	})
	if err != nil {
		return err
	}

	err = db.Update(func(txn *badger.Txn) error {
		updatedRecord.RegDTTM = existingRecord.RegDTTM
		updatedRecord.ID = existingRecord.ID
		value, _ := json.Marshal(updatedRecord)
		err = txn.Set([]byte(id), value)
		if err != nil {
			return err
		}

		// Update Bleve index
		return bleveIndex.Index(id, updatedRecord)
	})

	return err
}

func getRecords(queries []string, queryType string, startDate, endDate time.Time) ([]Record, map[string]Stat, map[string]Stat, float64, float64, float64, error) {
	var results []Record = []Record{}
	var stat map[string]Stat = map[string]Stat{}
	var statCredit map[string]Stat = map[string]Stat{}
	var totalPay float64 = 0
	var totalCreditPay float64 = 0
	var totalIncome float64 = 0

	boolQuery := bleve.NewBooleanQuery()

	// 기존 쿼리 조건 추가
	for _, query := range queries {
		matchQuery := bleve.NewMatchQuery(query)
		if queryType == "AND" {
			boolQuery.AddMust(matchQuery)
		} else {
			boolQuery.AddShould(matchQuery)
		}
	}

	dateRangeQuery := bleve.NewDateRangeQuery(startDate, endDate)
	dateRangeQuery.SetField("date") // 'date' 필드에 대해 날짜 범위 검색
	boolQuery.AddMust(dateRangeQuery)

	search := bleve.NewSearchRequest(boolQuery)

	// 페이징 리마크 - 일단 보류
	// search.Size = pageSize
	// search.From = (page - 1) * pageSize
	search.Size = 1000

	search.SortBy([]string{"date", "time", "_score"}) // SORT ASC
	// search.SortBy([]string{"-date", "-time", "-_score"}) // SORT DESC

	searchResults, err := bleveIndex.Search(search)
	if err != nil {
		return nil, nil, nil, 0, 0, 0, err
	}

	accounts, _ := getAccountListMAP()

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
			switch record.PayType {
			case "direct":
				totalPay += record.Amount

				amount := record.Amount
				if s, exist := stat[record.Category]; exist {
					amount = s.Amount + record.Amount
				}
				stat[record.Category] = Stat{Category: record.Category, Amount: amount}
			case "credit":
				repayDay, err1 := strconv.Atoi(accounts[record.AccountID].RepayDay)
				useDayFrom, err2 := strconv.Atoi(accounts[record.AccountID].UseDayFrom)
				useDayTo, err3 := strconv.Atoi(accounts[record.AccountID].UseDayTo)

				// If meet err, keep the type not repaid
				if err1 != nil || err2 != nil || err3 != nil {
					totalCreditPay += record.Amount

					amount := record.Amount
					if s, exist := statCredit[record.Category]; exist {
						amount = s.Amount + record.Amount
					}
					statCredit[record.Category] = Stat{Category: record.Category, Amount: amount}

					continue
				}

				repayDate, useDateFrom, useDateTo := getCreditDates(repayDay, useDayFrom, useDayTo, endDate)
				recordDate, _ := time.Parse("2006-01-02 15:04", record.Date+" "+record.Time)

				// Assume already paid: the day before "useDateFrom"
				if recordDate.Before(useDateFrom) {
					totalPay += record.Amount

					amount := record.Amount
					if s, exist := stat[record.Category]; exist {
						amount = s.Amount + record.Amount
					}
					stat[record.Category] = Stat{Category: record.Category, Amount: amount}

					continue
				}

				// Assume already paid: the day which meet all of the following conditions
				// * "recordDate" is Between "useDateFrom" and "useDateTo" - "useDateFrom" is already filtered by the above condition
				// * "endDate" is later than "repayDate"
				if (recordDate.Before(useDateTo) || recordDate.Equal(useDateTo)) && (endDate.After(repayDate) || endDate.Equal(repayDate)) {
					totalPay += record.Amount

					amount := record.Amount
					if s, exist := stat[record.Category]; exist {
						amount = s.Amount + record.Amount
					}
					stat[record.Category] = Stat{Category: record.Category, Amount: amount}

					continue
				}

				totalCreditPay += record.Amount

				amount := record.Amount
				if s, exist := statCredit[record.Category]; exist {
					amount = s.Amount + record.Amount
				}
				statCredit[record.Category] = Stat{Category: record.Category, Amount: amount}
			}
		case "record_type_income":
			totalIncome += record.Amount
		}
	}

	return results, stat, statCredit, totalPay, totalCreditPay, totalIncome, nil
}
