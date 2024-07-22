package server

// Payment account - cash, card. for input completion
type Account struct {
	ID          string `json:"id"`
	AccountName string `json:"account-name"`
	PayType     string `json:"pay-type"` // direct, credit, hybrid(revolving)
	RepayDay    string `json:"repay-day,omitempty"`
	UseDayFrom  string `json:"use-day-from,omitempty"`
	UseDayTo    string `json:"use-day-to,omitempty"`
	Description string `json:"description,omitempty"`
	RegDTTM     string
}

// Paymenr record
type Record struct {
	ID              string  `json:"id"`
	TransactionType string  `json:"transaction-type"` // payment(record_type_pay), income(record_type_income)
	AccountID       string  `json:"account-id"`
	PayType         string  `json:"pay-type"` // direct, credit
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category"`
	Description     string  `json:"description,omitempty"`
	Date            string  `json:"date"`
	Time            string  `json:"time"`
	RegDTTM         string
}
