package server

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

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

func validateAccount(account Account) error {
	if account.AccountName == "" {
		return fmt.Errorf("name is required")
	}
	if account.PayType == "" {
		return fmt.Errorf("pay-type is required")
	}

	return nil
}

func validateRecord(record Record) error {
	if record.TransactionType == "" {
		return fmt.Errorf("transaction-type is required")
	}
	if record.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	if record.PayType == "" {
		return fmt.Errorf("pay-type is required")
	}
	if record.Amount == 0 {
		return fmt.Errorf("amount is required and must be non-zero")
	}
	if record.Category == "" {
		return fmt.Errorf("category is required")
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
