package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

func setupDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	password := r.URL.Query().Get("password")
	if password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	if bleveIndex != nil {
		bleveIndex.Close()
		bleveIndex = nil
	}
	if db != nil {
		db.Close()
		db = nil
	}

	err := initBadgerDB(password)
	if err != nil {
		// log.Printf("Failed to initialize BadgerDB: %v", err)

		httpStatus := http.StatusInternalServerError
		if strings.Contains(err.Error(), "Encryption key mismatch") {
			httpStatus = http.StatusBadRequest
		}

		http.Error(w, "Failed to initialize database", httpStatus)
		return
	}

	err = initBleveIndex()
	if err != nil {
		// log.Printf("Failed to initialize Bleve index: %v", err)
		http.Error(w, "Failed to initialize search index", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func addAccountHandler(w http.ResponseWriter, r *http.Request) {
	var account Account

	err := json.NewDecoder(r.Body).Decode(&account)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = addAccount(account)
	if err != nil {
		// log.Printf("Failed to add account: %v", err)

		httpStatus := http.StatusInternalServerError
		if strings.Contains(err.Error(), "required") {
			httpStatus = http.StatusBadRequest
		}

		http.Error(w, "Failed to add account", httpStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	err := deleteAccount(recordID)
	if err != nil {
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("id")
	if accountID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	var updatedAccount Account
	err := json.NewDecoder(r.Body).Decode(&updatedAccount)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = validateAccount(updatedAccount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = updateAccount(accountID, updatedAccount)
	if err != nil {
		http.Error(w, "Failed to update record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getAccountListHandler(w http.ResponseWriter, r *http.Request) {
	accounts, err := getAccountList()
	if err != nil {
		http.Error(w, "Failed to get accounts", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accounts)
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
		// log.Printf("Failed to add record: %v", err)

		httpStatus := http.StatusInternalServerError
		if strings.Contains(err.Error(), "required") {
			httpStatus = http.StatusBadRequest
		}

		http.Error(w, "Failed to add record", httpStatus)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	accountID := r.URL.Query().Get("id")
	if accountID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	err := deleteRecord(accountID)
	if err != nil {
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func updateRecordHandler(w http.ResponseWriter, r *http.Request) {
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
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

	err = updateRecord(recordID, updatedRecord)
	if err != nil {
		http.Error(w, "Failed to update record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getRecordHandler(w http.ResponseWriter, r *http.Request) {
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
	// if pageSize < 1 || pageSize > 100 {
	if pageSize < 1 {
		pageSize = 10
	}

	records, sumPay, sumIncome, totalCount, err := getRecords(queries, page, pageSize, queryType)
	if err != nil {
		// log.Printf("Failed to search records: %v", err)
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

	w.WriteHeader(http.StatusOK)
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
		// log.Printf("Failed to get sum: %v", err)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
