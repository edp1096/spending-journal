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

// https://www.card-gorilla.com/contents/detail/2111
var CardDates = map[string][][]string{
	"롯데": {{"1", "18", "17"}, {"5", "22", "21"}, {"7", "24", "23"}, {"10", "27", "26"}, {"14", "1", "31"}, {"15", "2", "1"}, {"17", "4", "3"}, {"20", "7", "6"}, {"21", "8", "7"}, {"22", "9", "8"}, {"23", "10", "9"}, {"24", "11", "10"}, {"25", "12", "11"}},
	"삼성": {{"1", "20", "19"}, {"5", "24", "23"}, {"10", "29", "28"}, {"11", "31", "29"}, {"12", "31", "31"}, {"13", "1", "31"}, {"14", "1", "31"}, {"15", "3", "2"}, {"18", "6", "5"}, {"21", "9", "8"}, {"22", "10", "9"}, {"23", "11", "10"}, {"24", "12", "11"}, {"25", "13", "12"}, {"26", "14", "13"}},
	"신한": {{"1", "18", "17"}, {"2", "19", "18"}, {"3", "20", "19"}, {"4", "21", "20"}, {"5", "22", "21"}, {"6", "23", "22"}, {"7", "24", "23"}, {"8", "25", "24"}, {"9", "26", "25"}, {"10", "27", "26"}, {"11", "28", "27"}, {"12", "29", "28"}, {"13", "31", "29"}, {"14", "1", "31"}, {"15", "2", "1"}, {"16", "3", "2"}, {"17", "4", "3"}, {"18", "5", "4"}, {"19", "6", "5"}, {"20", "7", "6"}, {"21", "8", "7"}, {"22", "9", "8"}, {"23", "10", "9"}, {"24", "11", "10"}, {"25", "12", "11"}, {"26", "13", "12"}, {"27", "14", "13"}},
	"우리": {{"1", "18", "17"}, {"2", "19", "18"}, {"3", "20", "19"}, {"4", "21", "20"}, {"5", "22", "21"}, {"6", "23", "22"}, {"7", "24", "23"}, {"8", "25", "24"}, {"9", "26", "25"}, {"10", "27", "26"}, {"11", "28", "27"}, {"12", "29", "28"}, {"13", "31", "29"}, {"14", "1", "31"}, {"15", "2", "1"}, {"16", "3", "2"}, {"17", "4", "3"}, {"18", "5", "4"}, {"19", "6", "5"}, {"20", "7", "6"}, {"21", "8", "7"}, {"22", "9", "8"}, {"23", "10", "9"}, {"24", "11", "10"}, {"25", "12", "11"}, {"27", "14", "13"}},
	"하나": {{"1", "19", "18"}, {"5", "23", "22"}, {"7", "25", "24"}, {"8", "26", "25"}, {"10", "28", "27"}, {"12", "31", "29"}, {"13", "1", "31"}, {"14", "2", "1"}, {"15", "3", "2"}, {"17", "5", "4"}, {"18", "6", "5"}, {"20", "8", "7"}, {"21", "9", "8"}, {"23", "11", "10"}, {"25", "13", "12"}, {"27", "15", "14"}},
	"현대": {{"1", "20", "19"}, {"5", "24", "23"}, {"10", "29", "28"}, {"12", "1", "31"}, {"15", "4", "3"}, {"20", "9", "8"}, {"23", "12", "11"}, {"24", "13", "12"}, {"25", "14", "13"}, {"26", "15", "14"}},
	"기업": {{"1", "17", "16"}, {"2", "18", "17"}, {"3", "19", "18"}, {"4", "20", "19"}, {"5", "21", "20"}, {"6", "22", "21"}, {"7", "23", "22"}, {"8", "24", "23"}, {"9", "25", "24"}, {"10", "26", "25"}, {"11", "27", "26"}, {"12", "28", "27"}, {"13", "29", "28"}, {"14", "31", "29"}, {"15", "1", "31"}, {"16", "2", "1"}, {"17", "3", "2"}, {"18", "4", "3"}, {"19", "5", "4"}, {"20", "6", "5"}, {"21", "7", "6"}, {"22", "8", "7"}, {"23", "9", "8"}, {"24", "10", "9"}, {"25", "11", "10"}, {"26", "12", "11"}, {"27", "13", "12"}},
	"국민": {{"1", "18", "17"}, {"2", "19", "18"}, {"3", "20", "19"}, {"4", "21", "20"}, {"5", "22", "21"}, {"6", "23", "22"}, {"7", "24", "23"}, {"8", "25", "24"}, {"9", "26", "25"}, {"10", "27", "26"}, {"11", "28", "27"}, {"12", "29", "28"}, {"13", "31", "29"}, {"14", "1", "31"}, {"15", "2", "1"}, {"16", "3", "2"}, {"17", "4", "3"}, {"18", "5", "4"}, {"19", "6", "5"}, {"20", "7", "6"}, {"21", "8", "7"}, {"22", "9", "8"}, {"23", "10", "9"}, {"24", "11", "10"}, {"25", "12", "11"}, {"26", "13", "12"}, {"27", "14", "13"}},
	"농협": {{"1", "18", "17"}, {"2", "19", "18"}, {"3", "20", "19"}, {"4", "21", "20"}, {"5", "22", "21"}, {"6", "23", "22"}, {"7", "24", "23"}, {"8", "25", "24"}, {"9", "26", "25"}, {"10", "27", "26"}, {"11", "28", "27"}, {"12", "29", "28"}, {"13", "31", "29"}, {"14", "1", "31"}, {"15", "2", "1"}, {"16", "3", "2"}, {"17", "4", "3"}, {"18", "5", "4"}, {"19", "6", "5"}, {"20", "7", "6"}, {"21", "8", "7"}, {"22", "9", "8"}, {"23", "10", "9"}, {"24", "11", "10"}, {"25", "12", "11"}, {"26", "13", "12"}, {"27", "14", "13"}},
}

func getCreditPastMonthCount(repayDay, useDayFrom, useDayTo int) (int, int) {
	pointOfMonthNum := []int{-2, -1, 0}

	monthFromIDX := 0
	monthToIDX := 1
	if useDayFrom < repayDay {
		monthFromIDX++
		monthToIDX++
	}
	if useDayFrom < useDayTo {
		monthToIDX--
	}

	return pointOfMonthNum[monthFromIDX], pointOfMonthNum[monthToIDX]
}

func getCreditDates(repayDay, useDayFrom, useDayTo int, refernceDate time.Time) (repayDate, useDateFrom, useDateTo time.Time) {
	year, month, _ := refernceDate.Date()
	location := refernceDate.Location()

	useMonthFrom, useMonthTo := getCreditPastMonthCount(repayDay, useDayFrom, useDayTo)

	repayDate = time.Date(year, month, int(repayDay), 0, 0, 0, 0, location)

	useDateFrom = repayDate.AddDate(0, useMonthFrom, 0)
	useDateFrom = time.Date(useDateFrom.Year(), useDateFrom.Month(), int(useDayFrom), 0, 0, 0, 0, location)

	useDateTo = repayDate.AddDate(0, useMonthTo, 0)
	daysOfMonth := time.Date(useDateTo.Year(), useDateTo.Month()+1, 0, 0, 0, 0, 0, location).Day()
	if useDayTo > daysOfMonth {
		useDayTo = daysOfMonth
	}
	useDateTo = time.Date(useDateTo.Year(), useDateTo.Month(), int(useDayTo), 23, 59, 59, 0, location)

	return repayDate, useDateFrom, useDateTo
}
