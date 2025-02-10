package l402

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Wallet represents the minimum interface needed to handle L402 payments
type Wallet interface {
	// Pay processes the payment required by the L402
	Pay(response *L402Response) error
}

// HTTP402Client is an HTTP client that automatically handles L402 payment required responses
type HTTP402Client struct {
	httpClient *http.Client
	wallet     Wallet
}

// NewHTTP402Client creates a new L402 client with the given wallet
func NewHTTP402Client(wallet Wallet) *HTTP402Client {
	return &HTTP402Client{
		httpClient: http.DefaultClient,
		wallet:     wallet,
	}
}

// Do performs an HTTP request and automatically handles 402 Payment Required responses
func (c *HTTP402Client) Do(req *http.Request) (*http.Response, error) {
	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// If it's not a 402, return the response/error as is
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil
	}

	fmt.Printf("Received 402 Payment Required response\n")

	// Parse the L402 response
	response := &L402Response{}
	err = json.NewDecoder(resp.Body).Decode(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse L402 response: %w", err)
	}

	// Close the initial 402 response
	resp.Body.Close()

	// Process payment using wallet
	err = c.wallet.Pay(response)
	if err != nil {
		return nil, fmt.Errorf("failed to process L402 payment: %w", err)
	}

	// prompt user to pay
	var yn string
	fmt.Printf("Simulate the payment (visiting the URL) before continuing? (y/n): ")
	fmt.Scanln(&yn)
	if yn == "n" {
		return nil, fmt.Errorf("the client did not pay")
	}

	// After paying we should be able to access the resource
	return c.httpClient.Do(req)
}
