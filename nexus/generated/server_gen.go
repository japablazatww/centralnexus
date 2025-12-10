package generated

import (
	"encoding/json"
	"net/http"
	"github.com/japablazatww/libreria-a"
)



func RegisterHandlers(mux *http.ServeMux) {
	
	mux.HandleFunc("/liba/GetUserBalance", handleGetUserBalance)
	
	mux.HandleFunc("/liba/Transfer", handleTransfer)
	
	mux.HandleFunc("/liba/GetSystemStatus", handleGetSystemStatus)
	
}


func handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GetUserBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call underlying library
	res, err := liba.GetUserBalance(
		req.UserID,
		req.AccountID,
		
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
}

func handleTransfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call underlying library
	res, err := liba.Transfer(
		req.SourceAccount,
		req.DestAccount,
		req.Amount,
		req.Currency,
		
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
}

func handleGetSystemStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GetSystemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Call underlying library
	res, err := liba.GetSystemStatus(
		req.Code,
		
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
}

