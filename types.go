package server

// Payment method - cash, card. for input completion
type Method struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	PayMethod   string `json:"pay-method"` // cash, debit-card (check-card, debit-card), postpaid-card (credit-card)
	PayType     string `json:"pay-type"`   // direct-pay, credit-pay
	Description string `json:"description"`
	RegDTTM     string
}

// Paymenr record
type Record struct {
	ID              string  `json:"id"`
	TransactionType string  `json:"transaction-type"`
	PayMethod       string  `json:"pay-method"`
	Currency        string  `json:"currency"`
	Amount          float64 `json:"amount"`
	Category        string  `json:"category"`
	Description     string  `json:"description"`
	Date            string  `json:"date"`
	Time            string  `json:"time"`
	RegDTTM         string
}

type ErrorResponse struct {
	Error string `json:"error"`
}
