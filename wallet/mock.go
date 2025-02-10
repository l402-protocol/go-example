package wallet

import (
	"fmt"

	"github.com/l402-protocol/go-example/l402"
)

// MockWallet is a wallet that only logs payment requests but cannot actually pay
type MockWallet struct{}

// NewMockWallet creates a new mock wallet
func NewMockWallet() *MockWallet {
	return &MockWallet{}
}

// Pay implements the l402.Wallet interface but only logs the request and returns an error
func (w *MockWallet) Pay(response *l402.L402Response) error {
	fmt.Printf("Available offers:\n")
	for _, offer := range response.Offers {
		fmt.Printf("  Offer ID: %s\n", offer.ID)
		fmt.Printf("  Title: %s\n", offer.Title)
		fmt.Printf("  Price: $%.2f %s\n", float64(offer.Amount)/100, offer.Currency)
		fmt.Printf("  SupportedPayment Methods: %v\n", offer.PaymentMethods)
		fmt.Printf(" --------\n\n")
	}

	return fmt.Errorf("mock wallet cannot process payments, run client with --fake to simulate a payment")
}
