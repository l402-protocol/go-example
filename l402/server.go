package l402

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

func NewL402Response(url string, offers []Offer) L402Response {
	paymentContextToken := uuid.New().String()

	return L402Response{
		Version:             L402_VERSION,
		PaymentRequestURL:   url,
		PaymentContextToken: paymentContextToken,
		Offers:              offers,
		TermsURL:            "https://example.com/terms",
	}
}

func NewPayReqResponse(
	paymentRequest PaymentRequestRequest) (PaymentRequestResponse, error) {
	if paymentRequest.PaymentMethod != string(FakePay) {
		return PaymentRequestResponse{}, errors.New("payment method not supported")
	}

	payReq := PayReq{
		CheckoutURL: "http://localhost:8080/simulate-payment",
	}

	return PaymentRequestResponse{
		Version:        L402_VERSION,
		ExpiresAt:      time.Now().Add(time.Minute * 10).Format(time.RFC3339),
		PaymentRequest: payReq,
	}, nil
}
