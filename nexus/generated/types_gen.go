package generated


// GetUserBalanceRequest defines the input for GetUserBalance
type GetUserBalanceRequest struct {
	
	UserID string `json:"user_i_d"`
	
	AccountID string `json:"account_i_d"`
	
}

// TransferRequest defines the input for Transfer
type TransferRequest struct {
	
	SourceAccount string `json:"source_account"`
	
	DestAccount string `json:"dest_account"`
	
	Amount float64 `json:"amount"`
	
	Currency string `json:"currency"`
	
}

// GetSystemStatusRequest defines the input for GetSystemStatus
type GetSystemStatusRequest struct {
	
	Code string `json:"code"`
	
}

