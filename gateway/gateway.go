package gateway

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/l402-protocol/go-example/l402"
)

type Gateway struct {
	mux    *http.ServeMux
	logger *slog.Logger
}

func NewGateway() *Gateway {
	// Create a JSON logger with timestamp and caller info
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	g := &Gateway{
		mux:    http.NewServeMux(),
		logger: logger,
	}
	g.routes()
	return g
}

func (g *Gateway) routes() {
	g.mux.HandleFunc("POST /payment-request", g.handlePaymentRequest)
	g.mux.HandleFunc("POST /charge", g.handleCharge)
	g.mux.HandleFunc("GET /checkout", g.handleCheckout)
}

func (g *Gateway) handlePaymentRequest(w http.ResponseWriter, r *http.Request) {
	var req l402.PaymentRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.logger.Error("failed to decode payment request",
			"error", err,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	g.logger.Info("received payment request",
		"offer_id", req.OfferID,
		"payment_method", req.PaymentMethod,
		"payment_context", req.PaymentContextToken,
	)

	// For demo purposes, we'll just handle the fake payment method
	if req.PaymentMethod != string(l402.FakePay) {
		g.logger.Warn("unsupported payment method",
			"payment_method", req.PaymentMethod,
			"offer_id", req.OfferID,
		)
		http.Error(w, "only fake payments are supported in demo", http.StatusBadRequest)
		return
	}

	// Create checkout URL with payment context and offer ID
	checkoutURL := url.URL{
		Scheme: "http",
		Host:   "localhost:8081",
		Path:   "/checkout",
	}
	q := checkoutURL.Query()
	q.Set("payment_context_token", req.PaymentContextToken)
	q.Set("offer_id", req.OfferID)
	checkoutURL.RawQuery = q.Encode()

	// Return payment request response with checkout URL
	resp := l402.PaymentRequestResponse{
		Version:   l402.L402_VERSION,
		ExpiresAt: "2025-12-31T23:59:59Z",
		PaymentRequest: l402.PayReq{
			CheckoutURL: checkoutURL.String(),
		},
	}

	g.logger.Info("payment request processed",
		"offer_id", req.OfferID,
		"checkout_url", checkoutURL.String(),
		"expires_at", resp.ExpiresAt,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type ChargeRequest struct {
	Offers []l402.Offer `json:"offers"`
}

func (g *Gateway) handleCharge(w http.ResponseWriter, r *http.Request) {
	var req ChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		g.logger.Error("failed to decode charge request",
			"error", err,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Offers) == 0 {
		g.logger.Warn("no offers in charge request",
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "no offers provided", http.StatusBadRequest)
		return
	}

	g.logger.Info("received charge request",
		"num_offers", len(req.Offers),
		"offer_ids", getOfferIDs(req.Offers),
	)

	// In a real implementation, this would validate the payment
	// For demo purposes, we'll just return a successful L402 response with the received offers
	l402Response := l402.L402Response{
		Version:             l402.L402_VERSION,
		PaymentRequestURL:   "http://localhost:8081/payment-request",
		PaymentContextToken: "demo-token-123",
		Offers:              req.Offers,
	}

	g.logger.Info("charge request processed",
		"payment_context", l402Response.PaymentContextToken,
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(l402Response)
}

func (g *Gateway) handleCheckout(w http.ResponseWriter, r *http.Request) {
	// Get payment context and offer ID from URL parameters
	paymentContext := r.URL.Query().Get("payment_context_token")
	offerID := r.URL.Query().Get("offer_id")

	g.logger.Info("received checkout request",
		"payment_context", paymentContext,
		"offer_id", offerID,
		"remote_addr", r.RemoteAddr,
	)

	if paymentContext == "" || offerID == "" {
		g.logger.Warn("missing parameters in checkout request",
			"payment_context", paymentContext,
			"offer_id", offerID,
		)
		http.Error(w, "missing payment context or offer ID", http.StatusBadRequest)
		return
	}

	// Notify the backend server that payment was successful
	backendReq, err := http.NewRequest("POST", "http://localhost:8080/payment-success", nil)
	if err != nil {
		g.logger.Error("failed to create backend request",
			"error", err,
		)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// Add payment context and offer info
	backendReq.Header.Set("X-Payment-Context", paymentContext)
	backendReq.Header.Set("X-Offer-ID", offerID)

	start := time.Now()
	resp, err := http.DefaultClient.Do(backendReq)
	if err != nil {
		g.logger.Error("failed to notify backend",
			"error", err,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		http.Error(w, "failed to notify backend", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		g.logger.Error("backend failed to process payment",
			"status_code", resp.StatusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
		http.Error(w, "backend failed to process payment", http.StatusInternalServerError)
		return
	}

	g.logger.Info("payment processed successfully",
		"payment_context", paymentContext,
		"offer_id", offerID,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	// Return a nice HTML page indicating success
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<html>
			<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 40px auto; text-align: center;">
				<h1 style="color: #4CAF50;">Payment Successful!</h1>
				<p>Your payment for offer %s has been processed.</p>
				<p>You can now return to the application and access your content.</p>
			</body>
		</html>
	`, offerID)
}

// Helper function to extract offer IDs for logging
func getOfferIDs(offers []l402.Offer) []string {
	ids := make([]string, len(offers))
	for i, offer := range offers {
		ids[i] = offer.ID
	}
	return ids
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.mux.ServeHTTP(w, r)
}

// Start starts the gateway server on the specified port
func (g *Gateway) Start(port string) error {
	g.logger.Info("starting gateway server", "port", port)
	return http.ListenAndServe(port, g)
}
