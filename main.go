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
	"syscall"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/crypto/pbkdf2"
)

type Record struct {
	Type        string `json:"type"`
	Amount      int    `json:"amount"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Date        string `json:"date"`
	Time        string `json:"time"`
	RegDTTM     string `json:"regdttm"`
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

func addRecord(record Record) error {
	now := time.Now()
	timestamp := now.Unix()
	key := fmt.Sprintf("record:%d", timestamp)

	regdttm := now.Format("20060102150405")
	record.RegDTTM = regdttm

	err := db.Update(func(txn *badger.Txn) error {
		value, _ := json.Marshal(record)
		return txn.Set([]byte(key), value)
	})

	if err != nil {
		return err
	}

	return bleveIndex.Index(key, record)
}

func searchRecordsByDescription(searchTerm string) ([]Record, int, error) {
	var results []Record
	var totalSum int

	query := bleve.NewMatchQuery(searchTerm)
	search := bleve.NewSearchRequest(query)
	searchResults, err := bleveIndex.Search(search)
	if err != nil {
		return nil, 0, err
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
		totalSum += record.Amount
	}

	return results, totalSum, nil
}

func getSumByPeriod(startDate, endDate string) (int, error) {
	var sum int

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
					sum += record.Amount
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	return sum, err
}

func initializeHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	password := r.URL.Query().Get("password")

	err = initBadgerDB(password)
	if err != nil {
		// log.Fatalf("Failed to initialize BadgerDB: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = initBleveIndex()
	if err != nil {
		// log.Fatalf("Failed to initialize Bleve index: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func addRecordHandler(w http.ResponseWriter, r *http.Request) {
	var record Record

	err := json.NewDecoder(r.Body).Decode(&record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = addRecord(record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	records, _, err := searchRecordsByDescription(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(records)
}

func searchRecordWithSumHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	records, sum, err := searchRecordsByDescription(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := struct {
		Records  []Record `json:"records"`
		TotalSum int      `json:"total_sum"`
	}{
		Records:  records,
		TotalSum: sum,
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

	sum, err := getSumByPeriod(startDate, endDate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"sum": sum})
}

func main() {
	var err error

	mux := http.NewServeMux()
	mux.HandleFunc("GET /db/init", initializeHandler)
	mux.HandleFunc("POST /record", addRecordHandler)
	mux.HandleFunc("GET /record/search", searchRecordHandler)
	mux.HandleFunc("GET /record/search-with-sum", searchRecordWithSumHandler)
	mux.HandleFunc("GET /record/sum", getSumHandler)

	fmt.Println("Server starting on localhost:8080")
	server := &http.Server{Addr: "127.0.0.1:8080", Handler: mux}
	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println(err)
		}
	}()

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt, syscall.SIGTERM)

	<-kill

	if bleveIndex != nil {
		bleveIndex.Close()
	}
	if db != nil {
		db.Close()
	}

	err = server.Shutdown(context.Background())
	if err != nil {
		log.Printf("Shutdown error: %s", err)
	}
}
