package main

import (
	"fmt"

	"github.com/japablazatww/centralnexus/nexus/generated"
)

func main() {
	client := generated.NewClient("http://localhost:8080")

	// 1. Check System Status (using generic Params)
	fmt.Println("--- Testing GetSystemStatus ---")
	statusReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"code": "ADMIN123",
		},
	}
	// NOTICE: Using namespaced LibreriaA -> System
	status, err := client.LibreriaA.System.GetSystemStatus(statusReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("System Status: %v\n", status)
	}

	// 2. Get User Balance (Testing different Cases)
	fmt.Println("\n--- Testing GetUserBalance (CamelCase/SnakeCase check) ---")
	balanceReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"user_id":   "user_001", // Snake
			"AccountId": "acc_999",  // Pascal
		},
	}
	// NOTICE: LibreriaA -> Transfers -> National
	balance, err := client.LibreriaA.Transfers.National.GetUserBalance(balanceReq)
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
	// NOTICE: LibreriaA -> Transfers -> National
	transferRes, err := client.LibreriaA.Transfers.National.Transfer(transferReq)
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
	// NOTICE: LibreriaA -> Transfers -> International
	intRes, err := client.LibreriaA.Transfers.International.InternationalTransfer(intTransReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("International Transfer Result: %v\n", intRes)
	}
}
