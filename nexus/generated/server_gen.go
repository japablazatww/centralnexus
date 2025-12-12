package generated

import (
	"encoding/json"
	"net/http"
	"strings"
    
	
	libreria_a_system "github.com/japablazatww/libreria-a/system"
	
	libreria_a_transfers_international "github.com/japablazatww/libreria-a/transfers/international"
	
	libreria_a_transfers_national "github.com/japablazatww/libreria-a/transfers/national"
	
	libreria_b_loans "github.com/japablazatww/libreria-b/loans"
	
)

func RegisterHandlers(mux *http.ServeMux) {
	
	mux.HandleFunc("/libreria-a.system.GetSystemStatus", handlelibreria_a_system_GetSystemStatus)
	
	mux.HandleFunc("/libreria-a.transfers.national.GetUserBalance", handlelibreria_a_transfers_national_GetUserBalance)
	
	mux.HandleFunc("/libreria-a.transfers.national.Transfer", handlelibreria_a_transfers_national_Transfer)
	
	mux.HandleFunc("/libreria-a.transfers.national.ComplexTransfer", handlelibreria_a_transfers_national_ComplexTransfer)
	
	mux.HandleFunc("/libreria-a.transfers.international.InternationalTransfer", handlelibreria_a_transfers_international_InternationalTransfer)
	
	mux.HandleFunc("/libreria-b.loans.CalculateLoan", handlelibreria_b_loans_CalculateLoan)
	
}


func handlelibreria_a_system_GetSystemStatus(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_a_system_GetSystemStatus(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_a_system_GetSystemStatus(params map[string]interface{}) (interface{}, error) {
    // Inputs: code(string), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_code string
    

    // Fuzzy Match Logic
    found_code := false
    target_code := strings.ToLower(strings.ReplaceAll("code", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_code {
            
            val_code, _ = v.(string)
            
            found_code = true
            break
        }
    }
    
    if !found_code {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_a_system.GetSystemStatus(val_code, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

func handlelibreria_a_transfers_national_GetUserBalance(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_a_transfers_national_GetUserBalance(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_a_transfers_national_GetUserBalance(params map[string]interface{}) (interface{}, error) {
    // Inputs: user_i_d(string), account_i_d(string), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_user_i_d string
    

    // Fuzzy Match Logic
    found_user_i_d := false
    target_user_i_d := strings.ToLower(strings.ReplaceAll("user_i_d", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_user_i_d {
            
            val_user_i_d, _ = v.(string)
            
            found_user_i_d = true
            break
        }
    }
    
    if !found_user_i_d {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_account_i_d string
    

    // Fuzzy Match Logic
    found_account_i_d := false
    target_account_i_d := strings.ToLower(strings.ReplaceAll("account_i_d", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_account_i_d {
            
            val_account_i_d, _ = v.(string)
            
            found_account_i_d = true
            break
        }
    }
    
    if !found_account_i_d {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_national.GetUserBalance(val_user_i_d, val_account_i_d, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

func handlelibreria_a_transfers_national_Transfer(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_a_transfers_national_Transfer(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_a_transfers_national_Transfer(params map[string]interface{}) (interface{}, error) {
    // Inputs: source_account(string), dest_account(string), amount(float64), currency(string), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_source_account string
    

    // Fuzzy Match Logic
    found_source_account := false
    target_source_account := strings.ToLower(strings.ReplaceAll("source_account", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_source_account {
            
            val_source_account, _ = v.(string)
            
            found_source_account = true
            break
        }
    }
    
    if !found_source_account {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_dest_account string
    

    // Fuzzy Match Logic
    found_dest_account := false
    target_dest_account := strings.ToLower(strings.ReplaceAll("dest_account", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_dest_account {
            
            val_dest_account, _ = v.(string)
            
            found_dest_account = true
            break
        }
    }
    
    if !found_dest_account {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_amount float64
    

    // Fuzzy Match Logic
    found_amount := false
    target_amount := strings.ToLower(strings.ReplaceAll("amount", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_amount {
            
            val_amount, _ = v.(float64)
            
            found_amount = true
            break
        }
    }
    
    if !found_amount {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_currency string
    

    // Fuzzy Match Logic
    found_currency := false
    target_currency := strings.ToLower(strings.ReplaceAll("currency", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_currency {
            
            val_currency, _ = v.(string)
            
            found_currency = true
            break
        }
    }
    
    if !found_currency {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_national.Transfer(val_source_account, val_dest_account, val_amount, val_currency, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

func handlelibreria_a_transfers_national_ComplexTransfer(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_a_transfers_national_ComplexTransfer(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_a_transfers_national_ComplexTransfer(params map[string]interface{}) (interface{}, error) {
    // Inputs: req(TransferRequest), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_req libreria_a_transfers_national.TransferRequest
    

    // Fuzzy Match Logic
    found_req := false
    target_req := strings.ToLower(strings.ReplaceAll("req", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_req {
            
            // Complex Type: Convert map -> json -> struct
            jsonBody, _ := json.Marshal(v)
            json.Unmarshal(jsonBody, &val_req)
            
            found_req = true
            break
        }
    }
    
    if !found_req {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_national.ComplexTransfer(val_req, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

func handlelibreria_a_transfers_international_InternationalTransfer(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_a_transfers_international_InternationalTransfer(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_a_transfers_international_InternationalTransfer(params map[string]interface{}) (interface{}, error) {
    // Inputs: source_account(string), dest_iban(string), amount(float64), swift_code(string), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_source_account string
    

    // Fuzzy Match Logic
    found_source_account := false
    target_source_account := strings.ToLower(strings.ReplaceAll("source_account", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_source_account {
            
            val_source_account, _ = v.(string)
            
            found_source_account = true
            break
        }
    }
    
    if !found_source_account {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_dest_iban string
    

    // Fuzzy Match Logic
    found_dest_iban := false
    target_dest_iban := strings.ToLower(strings.ReplaceAll("dest_iban", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_dest_iban {
            
            val_dest_iban, _ = v.(string)
            
            found_dest_iban = true
            break
        }
    }
    
    if !found_dest_iban {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_amount float64
    

    // Fuzzy Match Logic
    found_amount := false
    target_amount := strings.ToLower(strings.ReplaceAll("amount", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_amount {
            
            val_amount, _ = v.(float64)
            
            found_amount = true
            break
        }
    }
    
    if !found_amount {
       // Optional: Log or Error if required param missing?
    }
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_swift_code string
    

    // Fuzzy Match Logic
    found_swift_code := false
    target_swift_code := strings.ToLower(strings.ReplaceAll("swift_code", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_swift_code {
            
            val_swift_code, _ = v.(string)
            
            found_swift_code = true
            break
        }
    }
    
    if !found_swift_code {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_international.InternationalTransfer(val_source_account, val_dest_iban, val_amount, val_swift_code, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

func handlelibreria_b_loans_CalculateLoan(w http.ResponseWriter, r *http.Request) {
	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 1. Extract Parameters
	params := req.Params
	
	// 2. Call Implementation
	resp, err := wrapperlibreria_b_loans_CalculateLoan(params)
	
	// 3. Response
	w.Header().Set("Content-Type", "application/json")
	
	if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
        return
	}
	json.NewEncoder(w).Encode(resp)
	
}

func wrapperlibreria_b_loans_CalculateLoan(params map[string]interface{}) (interface{}, error) {
    // Inputs: req(LoanRequest), 
    
    
    
    
    // Determine Type string (Primitive vs Complex)
    
    var val_req libreria_b_loans.LoanRequest
    

    // Fuzzy Match Logic
    found_req := false
    target_req := strings.ToLower(strings.ReplaceAll("req", "_", ""))
    
    for k, v := range params {
        normalizedK := strings.ToLower(strings.ReplaceAll(k, "_", ""))
        if normalizedK == target_req {
            
            // Complex Type: Convert map -> json -> struct
            jsonBody, _ := json.Marshal(v)
            json.Unmarshal(jsonBody, &val_req)
            
            found_req = true
            break
        }
    }
    
    if !found_req {
       // Optional: Log or Error if required param missing?
    }
    

    // Call
    ret0, ret1 := libreria_b_loans.CalculateLoan(val_req, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

