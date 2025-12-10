package generated

import (
	"encoding/json"
	"fmt"
	"net/http"
    "reflect"
    
	
	libreria_a_system "github.com/japablazatww/libreria-a/system"
	
	libreria_a_transfers_international "github.com/japablazatww/libreria-a/transfers/international"
	
	libreria_a_transfers_national "github.com/japablazatww/libreria-a/transfers/national"
	
)

func RegisterHandlers(mux *http.ServeMux) {
	
	mux.HandleFunc("/libreria-a.system.GetSystemStatus", handlelibreria_a_system_GetSystemStatus)
	
	mux.HandleFunc("/libreria-a.transfers.national.GetUserBalance", handlelibreria_a_transfers_national_GetUserBalance)
	
	mux.HandleFunc("/libreria-a.transfers.national.Transfer", handlelibreria_a_transfers_national_Transfer)
	
	mux.HandleFunc("/libreria-a.transfers.international.InternationalTransfer", handlelibreria_a_transfers_international_InternationalTransfer)
	
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
    
    
    var val_code string // simplified extraction
    if v, ok := params["code"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_code, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
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
    // Inputs: user_id(string), account_id(string), 
    
    
    var val_user_id string // simplified extraction
    if v, ok := params["user_id"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_user_id, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_account_id string // simplified extraction
    if v, ok := params["account_id"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_account_id, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_national.GetUserBalance(val_user_id, val_account_id, )
    
    
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
    
    
    var val_source_account string // simplified extraction
    if v, ok := params["source_account"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_source_account, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_dest_account string // simplified extraction
    if v, ok := params["dest_account"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_dest_account, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_amount float64 // simplified extraction
    if v, ok := params["amount"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_amount, _ = v.(float64)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_currency string // simplified extraction
    if v, ok := params["currency"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_currency, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_national.Transfer(val_source_account, val_dest_account, val_amount, val_currency, )
    
    
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
    
    
    var val_source_account string // simplified extraction
    if v, ok := params["source_account"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_source_account, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_dest_iban string // simplified extraction
    if v, ok := params["dest_iban"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_dest_iban, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_amount float64 // simplified extraction
    if v, ok := params["amount"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_amount, _ = v.(float64)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    
    var val_swift_code string // simplified extraction
    if v, ok := params["swift_code"]; ok {
        // Simple type assertion for PoC (float64 for json numbers)
        // In real world, use reflection or sophisticated casting
        // Here we assume happy path or simple cast
        // JSON numbers are float64.
        _ = v
        
        val_swift_code, _ = v.(string)
        
        
        // Dynamic fuzzy match fallback (omitted for brevity in this step, using direct key)
    }
    

    // Call
    ret0, ret1 := libreria_a_transfers_international.InternationalTransfer(val_source_account, val_dest_iban, val_amount, val_swift_code, )
    
    
    // Handle error convention (last return is error)
    if ret1 != nil {
        return nil, ret1
    }
    return ret0, nil
    
}

