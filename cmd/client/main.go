package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"

	"github.com/l402-protocol/go-example/l402"
	"github.com/l402-protocol/go-example/wallet"
)

const (
	authToken = "01badad5-f2f0-43cf-be44-4f8e1c4d8641"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	var (
		offerID = flag.String("offer-id", "offer_0001", "Offer ID that we will purchase")
		useFake = flag.Bool("fake", false, "Simulate a fake payment")
	)
	flag.Parse()

	// Create the appropriate wallet based on the --fake flag
	var w l402.Wallet
	if *useFake {
		w = wallet.NewFakeWallet(*offerID)
	} else {
		w = wallet.NewMockWallet()
	}

	// Create the L402 client with the selected wallet
	client := l402.NewHTTP402Client(w)

	// Make a request to the protected resource
	req, err := http.NewRequest("GET", "http://localhost:8080/private-resource", nil)
	if err != nil {
		logger.Error("failed to create request", "error", err)
		os.Exit(1)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+authToken)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("failed to execute request",
			"error", err,
			"url", req.URL.String(),
		)
		os.Exit(1)
	}
	defer resp.Body.Close()

	logger.Info("request completed",
		"status", resp.StatusCode,
		"url", req.URL.String(),
		"offer_id", *offerID,
	)
}
