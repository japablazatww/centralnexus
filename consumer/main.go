package main

import (
	"fmt"

	"github.com/japablazatww/centralnexus/nexus/generated"
)

func main() {
	client := generated.NewClient("http://localhost:8080")

	// 1. Check System Status
	fmt.Println("--- Testing GetSystemStatus ---")
	statusReq := generated.GetSystemStatusRequest{
		Code: "ADMIN123",
	}
	status, err := client.GetSystemStatus(statusReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("System Status: %v\n", status)
	}

	// 2. Get User Balance
	fmt.Println("\n--- Testing GetUserBalance ---")
	balanceReq := generated.GetUserBalanceRequest{
		UserID:    "user_001",
		AccountID: "acc_999",
	}
	balance, err := client.GetUserBalance(balanceReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Balance: %v\n", balance)
	}

	// 3. Transfer
	fmt.Println("\n--- Testing Transfer ---")
	transferReq := generated.TransferRequest{
		SourceAccount: "acc_999",
		DestAccount:   "acc_888",
		Amount:        50.0,
		Currency:      "GTQ",
	}
	txID, err := client.Transfer(transferReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Transfer ID: %v\n", txID)
	}

	// 4. Test Parameter Mapping (Case insensitivity check simulation)
	// The generated SDK uses strict JSON tags, but the Nexus server *could* benefit from
	// a more flexible decoder in a real scenario. For this PoC, we demonstrate
	// the standard generated client usage which guarantees contract compliance.
}
