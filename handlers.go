package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v3"
)

func initializeHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	if password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	err := initBadgerDB(password)
	if err != nil {
		log.Printf("Failed to initialize BadgerDB: %v", err)
		http.Error(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}

	err = initBleveIndex()
	if err != nil {
		log.Printf("Failed to initialize Bleve index: %v", err)
		http.Error(w, "Failed to initialize search index", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func addRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record Record

	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = addRecord(record)
	if err != nil {
		log.Printf("Failed to add record: %v", err)
		if strings.Contains(err.Error(), "required") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to add record", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func searchRecordHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	queries := strings.Fields(query)

	queryType := r.URL.Query().Get("queryType")
	if queryType != "AND" && queryType != "OR" {
		queryType = "OR"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	records, sumPay, sumIncome, totalCount, err := searchRecordsByDescription(queries, page, pageSize, queryType)
	if err != nil {
		log.Printf("Failed to search records: %v", err)
		http.Error(w, "Failed to search records", http.StatusInternalServerError)
		return
	}

	response := struct {
		Records    []Record `json:"records"`
		SumPay     float64  `json:"sum-pay"`
		SumIncome  float64  `json:"sum-income"`
		TotalCount int      `json:"total-count"`
		Page       int      `json:"page"`
		PageSize   int      `json:"page_size"`
	}{
		Records:    records,
		SumPay:     sumPay,
		SumIncome:  sumIncome,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	json.NewEncoder(w).Encode(response)
}

func getSumHandler(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")
	if startDate == "" || endDate == "" {
		http.Error(w, "Both 'start' and 'end' query parameters are required", http.StatusBadRequest)
		return
	}

	records, sumPay, sumIncome, err := getSumByPeriod(startDate, endDate)
	if err != nil {
		log.Printf("Failed to get sum: %v", err)
		http.Error(w, "Failed to calculate sum", http.StatusInternalServerError)
		return
	}

	response := struct {
		Records   []Record `json:"records"`
		SumPay    float64  `json:"sum-pay"`
		SumIncome float64  `json:"sum-income"`
	}{
		Records:   records,
		SumPay:    sumPay,
		SumIncome: sumIncome,
	}

	json.NewEncoder(w).Encode(response)
}

func deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		http.Error(w, "Query parameter 'id' is required", http.StatusBadRequest)
		return
	}

	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(recordID))
		if err != nil {
			return err
		}
		return bleveIndex.Delete(recordID)
	})
	if err != nil {
		log.Printf("Failed to delete record: %v", err)
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func updateRecordHandler(w http.ResponseWriter, r *http.Request) {
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		http.Error(w, "Query parameter 'id' is required", http.StatusBadRequest)
		return
	}

	var updatedRecord Record
	err := json.NewDecoder(r.Body).Decode(&updatedRecord)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = validateRecord(updatedRecord)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingRecord Record
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(recordID))
		if err != nil {
			return err
		}
		return item.Value(func(v []byte) error {
			return json.Unmarshal(v, &existingRecord)
		})
	})
	if err != nil {
		log.Printf("Failed to fetch record for update: %v", err)
		http.Error(w, "Record not found", http.StatusNotFound)
		return
	}

	err = db.Update(func(txn *badger.Txn) error {
		updatedRecord.RegDTTM = existingRecord.RegDTTM
		updatedRecord.ID = existingRecord.ID
		value, _ := json.Marshal(updatedRecord)
		err = txn.Set([]byte(recordID), value)
		if err != nil {
			return err
		}

		return bleveIndex.Index(recordID, updatedRecord)
	})
	if err != nil {
		log.Printf("Failed to update record: %v", err)
		http.Error(w, "Failed to update record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
