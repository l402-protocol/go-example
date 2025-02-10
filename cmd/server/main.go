package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/l402-protocol/go-example/l402"
)

var (
	offers []l402.Offer = []l402.Offer{
		{
			ID:             "offer_0001",
			Title:          "Pro",
			Description:    "10 credits for all the datasets in the platform",
			Type:           l402.TopUp,
			Amount:         10,
			Currency:       "USD",
			PaymentMethods: []l402.PaymentMethods{l402.Lightning, l402.FakePay},
		}, {
			ID:             "offer_0002",
			Title:          "Dataset purchase",
			Description:    "Unlimited access to a specific dataset",
			Type:           l402.OneTime,
			Amount:         100,
			Currency:       "USD",
			PaymentMethods: []l402.PaymentMethods{l402.Lightning, l402.Onchain, l402.FakePay},
		}, {
			ID:             "offer_0003",
			Title:          "Unlimited",
			Description:    "Unlimited access to all the datasets in the platform",
			Type:           l402.Subscription,
			Amount:         500,
			Currency:       "USD",
			PaymentMethods: []l402.PaymentMethods{l402.Lightning, l402.Onchain, l402.CreditCard, l402.FakePay},
		},
	}

	authToken = "01badad5-f2f0-43cf-be44-4f8e1c4d8641"
)

type Server struct {
	mux     *http.ServeMux
	hasPaid bool
	logger  *slog.Logger
}

func NewServer() *Server {
	// Create a JSON logger with timestamp and caller info
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s := &Server{
		mux:     http.NewServeMux(),
		hasPaid: false,
		logger:  logger,
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /private-resource", s.handleGetPrivateResource)
	s.mux.HandleFunc("POST /payment-success", s.handlePaymentSuccess) // Webhook from gateway
}

func (s *Server) handlePaymentSuccess(w http.ResponseWriter, r *http.Request) {
	// Verify payment context and offer ID from headers
	paymentContext := r.Header.Get("X-Payment-Context")
	offerID := r.Header.Get("X-Offer-ID")

	if paymentContext == "" || offerID == "" {
		s.logger.Error("missing payment context or offer ID in headers",
			"payment_context", paymentContext,
			"offer_id", offerID,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "missing payment context or offer ID", http.StatusBadRequest)
		return
	}

	// In a real implementation, we would:
	// 1. Verify the payment context is valid
	// 2. Check if the offer exists and is valid
	// 3. Update user's credits/access based on the offer type
	// For demo, we just set hasPaid to true
	s.hasPaid = true

	s.logger.Info("payment processed successfully",
		"payment_context", paymentContext,
		"offer_id", offerID,
		"remote_addr", r.RemoteAddr,
	)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleGetPrivateResource(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		s.logger.Warn("unauthorized request: missing bearer token",
			"remote_addr", r.RemoteAddr,
			"path", r.URL.Path,
		)
		http.Error(w, "Unauthorized: missing bearer token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != authToken {
		s.logger.Warn("unauthorized request: invalid token",
			"remote_addr", r.RemoteAddr,
			"path", r.URL.Path,
		)
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	s.logger.Debug("authenticated request received",
		"remote_addr", r.RemoteAddr,
		"path", r.URL.Path,
		"has_paid", s.hasPaid,
	)

	if !s.hasPaid {
		// Send offers to gateway's charge endpoint
		chargeReq := struct {
			Offers []l402.Offer `json:"offers"`
		}{
			Offers: offers,
		}

		body, err := json.Marshal(chargeReq)
		if err != nil {
			s.logger.Error("failed to marshal charge request",
				"error", err,
				"remote_addr", r.RemoteAddr,
			)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		s.logger.Debug("sending charge request to gateway",
			"num_offers", len(offers),
			"remote_addr", r.RemoteAddr,
		)

		resp, err := http.Post("http://localhost:8081/charge", "application/json", bytes.NewBuffer(body))
		if err != nil {
			s.logger.Error("failed to contact gateway",
				"error", err,
				"remote_addr", r.RemoteAddr,
			)
			http.Error(w, "gateway error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Forward the gateway's response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPaymentRequired)
		io.Copy(w, resp.Body)

		s.logger.Info("payment required response sent",
			"remote_addr", r.RemoteAddr,
			"status", http.StatusPaymentRequired,
		)
		return
	}

	// User has paid, return success response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "Access granted to dataset 123",
		"status":  "success",
	}
	json.NewEncoder(w).Encode(response)

	s.logger.Info("access granted to protected resource",
		"remote_addr", r.RemoteAddr,
		"path", r.URL.Path,
	)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func main() {
	server := NewServer()
	server.logger.Info("starting backend server", "port", ":8080")
	if err := http.ListenAndServe(":8080", server); err != nil {
		server.logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
}
