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
	// NOTICE: Using namespaced LibreriaA
	status, err := client.LibreriaA.GetSystemStatus(statusReq)
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
	balance, err := client.LibreriaA.GetUserBalance(balanceReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Balance: %v\n", balance)
	}

	// 3. Transfer
	fmt.Println("\n--- Testing Transfer ---")
	transferReq := generated.GenericRequest{
		Params: map[string]interface{}{
			"sourceAccount": "acc_999",
			"destAccount":   "acc_888",
			"amount":        50.0,
			"currency":      "GTQ",
		},
	}
	txID, err := client.LibreriaA.Transfer(transferReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Transfer ID: %v\n", txID)
	}
}
