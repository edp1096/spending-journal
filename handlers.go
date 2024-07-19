package server

import (
	"encoding/json"
	"log"
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

func addMethodHandler(w http.ResponseWriter, r *http.Request) {
	var method Method

	err := json.NewDecoder(r.Body).Decode(&method)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = addMethod(method)
	if err != nil {
		log.Printf("Failed to add method: %v", err)
		if strings.Contains(err.Error(), "required") {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Error(w, "Failed to add method", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func deleteMethodHandler(w http.ResponseWriter, r *http.Request) {
	recordID := r.URL.Query().Get("id")
	if recordID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	err := deleteMethod(recordID)
	if err != nil {
		http.Error(w, "Failed to delete record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func updateMethodHandler(w http.ResponseWriter, r *http.Request) {
	methodID := r.URL.Query().Get("id")
	if methodID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	var updatedMethod Method
	err := json.NewDecoder(r.Body).Decode(&updatedMethod)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = validateMethod(updatedMethod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = updateMethod(methodID, updatedMethod)
	if err != nil {
		http.Error(w, "Failed to update record", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func getMethodListHandler(w http.ResponseWriter, r *http.Request) {
	methods, err := getMethodList()
	if err != nil {
		http.Error(w, "Failed to get methods", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(methods)
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

func deleteRecordHandler(w http.ResponseWriter, r *http.Request) {
	methodID := r.URL.Query().Get("id")
	if methodID == "" {
		http.Error(w, "'id' is required", http.StatusBadRequest)
		return
	}

	err := deleteRecord(methodID)
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

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
