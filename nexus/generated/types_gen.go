package generated

// --- Shared Types ---

type GenericRequest struct {
	Params map[string]interface{} `json:"params"`
}


type TransferRequest struct {

    SourceAccount string `json:"source_account"`

    DestAccount string `json:"dest_account"`

    Amount float64 `json:"amount"`

    Currency string `json:"currency"`

}

type TransferResponse struct {

    TransactionID string `json:"transaction_id"`

    Status string `json:"status"`

}

type LoanRequest struct {

    Amount float64 `json:"amount"`

    Term int `json:"term"`

    UserType string `json:"user_type"`

}

type LoanResponse struct {

    Approved bool `json:"approved"`

    InterestRate float64 `json:"interest_rate"`

    MonthlyPay float64 `json:"monthly_pay"`

    Message string `json:"message"`

}

