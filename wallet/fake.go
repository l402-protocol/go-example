package wallet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/l402-protocol/go-example/l402"
)

// FakeWallet can be used to simulate a payment.
type FakeWallet struct {
	offerID string
	logger  *slog.Logger
}

// NewFakeWallet creates a new fake wallet
func NewFakeWallet(offerID string) *FakeWallet {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	return &FakeWallet{
		offerID: offerID,
		logger:  logger,
	}
}

// Pay simulates a payment by making the payment request
func (w *FakeWallet) Pay(response *l402.L402Response) error {
	var offer *l402.Offer
	for _, o := range response.Offers {
		if o.ID == w.offerID {
			offer = &o
			break
		}
	}

	if offer == nil {
		return fmt.Errorf("offer %s not found", w.offerID)
	}

	w.logger.Info("processing payment",
		"offer_id", offer.ID,
		"title", offer.Title,
		"amount", offer.Amount,
		"currency", offer.Currency,
	)

	// Create payment request
	payReq := l402.PaymentRequestRequest{
		OfferID:             offer.ID,
		PaymentMethod:       string(l402.FakePay),
		PaymentContextToken: response.PaymentContextToken,
	}

	body, err := json.Marshal(payReq)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request: %w", err)
	}

	// Make payment request to gateway
	resp, err := http.Post(response.PaymentRequestURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to make payment request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("payment request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse payment request response
	var payResp l402.PaymentRequestResponse
	if err := json.NewDecoder(resp.Body).Decode(&payResp); err != nil {
		return fmt.Errorf("failed to decode payment response: %w", err)
	}

	w.logger.Info("received payment request response",
		"checkout_url", payResp.PaymentRequest.CheckoutURL,
		"expires_at", payResp.ExpiresAt,
	)

	fmt.Printf("\nTo complete the payment, visit:\n%s\n", payResp.PaymentRequest.CheckoutURL)
	return nil
}
