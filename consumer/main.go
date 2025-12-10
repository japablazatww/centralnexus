package main

import (
	"fmt"
	"os"

	"github.com/japablazatww/centralnexus/nexus/generated"
)

func main() {
	baseURL := os.Getenv("NEXUS_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client := generated.NewClient(baseURL)

	// 1. Check System Status (using generic Params)
	fmt.Println("--- Testing GetSystemStatus ---")
	statusReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"code": "ADMIN123",
		},
	}
	// NOTICE: Using namespaced Libreriaa -> System
	status, err := client.Libreriaa.System.GetSystemStatus(statusReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("System Status: %v\n", status)
	}

	// 2. Get User Balance (Testing different Cases)
	fmt.Println("\n--- Testing GetUserBalance (CamelCase/SnakeCase check) ---")
	balanceReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"user_id":    "user_001", // Protocol enforces snake_case
			"account_id": "acc_999",  // Protocol enforces snake_case
		},
	}
	// NOTICE: Libreriaa -> Transfers -> National
	balance, err := client.Libreriaa.Transfers.National.GetUserBalance(balanceReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Balance: %v\n", balance)
	}

	// 3. Transfer
	fmt.Println("\n--- Testing Transfer (National) ---")
	transferReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"sourceAccount": "acc_999",
			"destAccount":   "acc_888",
			"amount":        50.0,
			"currency":      "GTQ",
		},
	}
	// NOTICE: Libreriaa -> Transfers -> National
	transferRes, err := client.Libreriaa.Transfers.National.Transfer(transferReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Transfer Result: %v\n", transferRes)
	}

	// 4. International Transfer
	fmt.Println("\n--- Testing International Transfer ---")
	intTransReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"source_account": "acc_999",
			"dest_iban":      "US123456789",
			"swift_code":     "SWIFT123",
			"amount":         2000.00,
		},
	}
	// NOTICE: Libreriaa -> Transfers -> International
	intRes, err := client.Libreriaa.Transfers.International.InternationalTransfer(intTransReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("International Transfer Result: %v\n", intRes)
	}

	// 5. Complex Transfer (Struct Input/Output)
	fmt.Println("\n--- Testing ComplexTransfer (Struct) ---")
	complexReqPayload := map[string]interface{}{
		"source_account": "ACC-COMPLEX-001",
		"dest_account":   "ACC-COMPLEX-002",
		"amount":         999.99,
		"currency":       "USD",
	}
	complexReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"req": complexReqPayload,
		},
	}
	complexRes, err := client.Libreriaa.Transfers.National.ComplexTransfer(complexReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Complex Transfer Result: %v\n", complexRes)
	}
}
