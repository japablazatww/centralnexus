package generated

import (
	"encoding/json"
	"net/http"
	"github.com/japablazatww/libreria-a"
	"fmt"
	"strings"
	"unicode"
)

func RegisterHandlers(mux *http.ServeMux) {
	
	mux.HandleFunc("/liba/GetUserBalance", handleGetUserBalance)
	
	mux.HandleFunc("/liba/Transfer", handleTransfer)
	
	mux.HandleFunc("/liba/GetSystemStatus", handleGetSystemStatus)
	
}

func getParam(params map[string]interface{}, name string) (interface{}, error) {
	// 1. Try exact match
	if v, ok := params[name]; ok { return v, nil }

	// 2. Case-Insensitive Match
	// Create a normalized map where keys are lowercased (without underscores for fuzzy matching might be better, but let's stick to lower case first)
	// For performance in a real app this should be done once per request, but for PoC this function is fine.
	target := strings.ToLower(name)
	targetNoUnderscore := strings.ReplaceAll(target, "_", "")

	for k, v := range params {
		kLower := strings.ToLower(k)
		if kLower == target { return v, nil }
		
		// 3. Fuzzy match (ignoring underscores) e.g. "user_id" vs "userid"
		kNoUnderscore := strings.ReplaceAll(kLower, "_", "")
		if kNoUnderscore == targetNoUnderscore { return v, nil }
	}

	return nil, fmt.Errorf("param %s not found in request params", name)
}

func toSnakeCase(str string) string {
	var matchFirstCap = unicode.IsUpper
	var result strings.Builder
	for i, r := range str {
		if matchFirstCap(r) && i > 0 {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

func toPascalCase(str string) string {
	if len(str) == 0 { return "" }
	return strings.ToUpper(str[:1]) + str[1:]
}




func handleGetUserBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := req.Params
	if params == nil {
		params = make(map[string]interface{})
	}

	// Dynamic Parameter Extraction
	
	val_userID, err := getParam(params, "userID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_userID string
	
	switch v := val_userID.(type) {
	case string:
		arg_userID = v
	
	
	default:
		_ = v
	}
	
	val_accountID, err := getParam(params, "accountID")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_accountID string
	
	switch v := val_accountID.(type) {
	case string:
		arg_accountID = v
	
	
	default:
		_ = v
	}
	

	// Call underlying library
	res, err := liba.GetUserBalance(
		arg_userID,
		arg_accountID,
		
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

	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := req.Params
	if params == nil {
		params = make(map[string]interface{})
	}

	// Dynamic Parameter Extraction
	
	val_sourceAccount, err := getParam(params, "sourceAccount")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_sourceAccount string
	
	switch v := val_sourceAccount.(type) {
	case string:
		arg_sourceAccount = v
	
	
	default:
		_ = v
	}
	
	val_destAccount, err := getParam(params, "destAccount")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_destAccount string
	
	switch v := val_destAccount.(type) {
	case string:
		arg_destAccount = v
	
	
	default:
		_ = v
	}
	
	val_amount, err := getParam(params, "amount")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_amount float64
	
	switch v := val_amount.(type) {
	case float64:
		arg_amount = v
	
	
	case string:
		// Try to handle string if needed, currently empty for strict types but avoided duplicate case
	
	default:
		_ = v
	}
	
	val_currency, err := getParam(params, "currency")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_currency string
	
	switch v := val_currency.(type) {
	case string:
		arg_currency = v
	
	
	default:
		_ = v
	}
	

	// Call underlying library
	res, err := liba.Transfer(
		arg_sourceAccount,
		arg_destAccount,
		arg_amount,
		arg_currency,
		
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

	var req GenericRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := req.Params
	if params == nil {
		params = make(map[string]interface{})
	}

	// Dynamic Parameter Extraction
	
	val_code, err := getParam(params, "code")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// Type Assertion/Conversion (Simplified for PoC - assumes correct JSON types or simple string conversions)
	var arg_code string
	
	switch v := val_code.(type) {
	case string:
		arg_code = v
	
	
	default:
		_ = v
	}
	

	// Call underlying library
	res, err := liba.GetSystemStatus(
		arg_code,
		
	)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"result": res})
}

