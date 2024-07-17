package main // import "app-server"

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/crypto/pbkdf2"
)

type Record struct {
	ID          string  `json:"id"`
	TradeType   string  `json:"trade-type"`
	PayMethod   string  `json:"pay-method"`
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Time        string  `json:"time"`
	RegDTTM     string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var db *badger.DB
var bleveIndex bleve.Index

func generateKey(password string, salt []byte) []byte {
	key := pbkdf2.Key([]byte(password), salt, 10000, 32, sha256.New)
	return key
}

func getSalt(filename string) ([]byte, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			return nil, err
		}
		if err := os.WriteFile(filename, salt, 0600); err != nil {
			return nil, err
		}
		return salt, nil
	}

	salt, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

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

func validateRecord(record Record) error {
	if record.TradeType == "" {
		return fmt.Errorf("trade-type is required")
	}
	if record.PayMethod == "" {
		return fmt.Errorf("pay-method is required")
	}
	if record.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if record.Amount == 0 {
		return fmt.Errorf("amount is required and must be non-zero")
	}
	if record.Category == "" {
		return fmt.Errorf("category is required")
	}
	if record.Description == "" {
		return fmt.Errorf("description is required")
	}
	if record.Date == "" {
		return fmt.Errorf("date is required")
	}
	if _, err := time.Parse("2006-01-02", record.Date); err != nil {
		return fmt.Errorf("invalid date format: use YYYY-MM-DD")
	}
	if record.Time != "" {
		if _, err := time.Parse("15:04", record.Time); err != nil {
			return fmt.Errorf("invalid time format: use HH:MM")
		}
	}
	return nil
}

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

		switch record.TradeType {
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

					switch record.TradeType {
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /setup/db", initializeHandler)
	mux.HandleFunc("POST /record", addRecordHandler)
	mux.HandleFunc("GET /record/search", searchRecordHandler)
	mux.HandleFunc("GET /record/sum", getSumHandler)
	mux.HandleFunc("DELETE /record/delete", deleteRecordHandler)
	mux.HandleFunc("PUT /record/update", updateRecordHandler)

	server := &http.Server{Addr: "127.0.0.1:8080", Handler: mux}
	go func() {
		fmt.Println("Server starting on localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if bleveIndex != nil {
		bleveIndex.Close()
	}
	if db != nil {
		db.Close()
	}

	fmt.Println("Server exited")
}
