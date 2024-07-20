package server

// Payment method - cash, card. for input completion
type Method struct {
	ID          string `json:"id"`
	AccountName string `json:"account-name"`
	PayType     string `json:"pay-type"` // direct, credit, hybrid(revolving)
	RepayDate   string `json:"repay-date,omitempty"`
	Description string `json:"description,omitempty"`
	RegDTTM     string
}

// Paymenr record
type Record struct {
	ID              string  `json:"id"`
	TransactionType string  `json:"transaction-type"` // payment, income
	PayMethod       string  `json:"pay-method"`
	RepayDate       string  `json:"repay-date,omitempty"`
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category"`
	Description     string  `json:"description,omitempty"`
	Date            string  `json:"date"`
	Time            string  `json:"time"`
	RegDTTM         string
}
