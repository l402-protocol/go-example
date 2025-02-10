package l402

const (
	L402_VERSION = "0.2.2"
)

type PaymentMethods string

// Payment methods can be added without breaking the spec.
// All payments happen outband. The L402 is just an HTTP flow to make sure the user gets the
// right payment details.
const (
	Lightning  PaymentMethods = "lightning"
	Onchain    PaymentMethods = "onchain"
	CreditCard PaymentMethods = "credit_card"

	// FakePay is a fake payment method to test the flow in this
	// demo.
	FakePay PaymentMethods = "fake-pay"
)

type OfferType string

// There may be other types if needed.
const (
	// OneTime is a one-time payment for a service or product.
	// Good use cases are ecommerce (flights, carts, domains...)
	OneTime OfferType = "one-time"

	// TopUp is a one-time payment to add funds to a balance.
	// Good for credit-based services (metered usage...)
	TopUp OfferType = "top-up"

	// Subscription is a recurring payment for a service or product.
	// Good use cases are recurring payments (saas...)
	Subscription OfferType = "subscription"
)

type Offer struct {
	// The ID of the offer.
	ID string `json:"id"`
	// A short title of what service/product is being purchased.
	Title string `json:"title"`
	// A description of the service/product.
	Description string `json:"description"`

	// Offer type.
	Type OfferType `json:"type"`

	// Balance is only used for TopUp offers. It shows how many credits will be added.
	Balance int `json:"balance"`

	// Amount is the amount to be paid. Currently using the smallest unit of the currency.
	Amount int `json:"amount"`

	// Currency of the amount. (USD, EUR, BTC, etc)
	Currency string `json:"currency"`

	// Payment methods supported by the offer.
	// How the client can pay this offer.
	PaymentMethods []PaymentMethods `json:"payment_methods"`
}

type L402Response struct {
	// Version
	Version string `json:"version"`

	// URL where the client can get payment request data for a specific offer/payment method
	PaymentRequestURL string `json:"payment_request_url"`

	// Unique identifier to link a payment
	PaymentContextToken string `json:"payment_context_token"`

	// Offers
	Offers []Offer `json:"offer"`

	// Terms and conditions
	TermsURL string `json:"terms_url"`
}

type PaymentRequestRequest struct {
	// OfferID the client is interested in.
	OfferID string `json:"offer_id"`
	// For what payment method is requesting the payment details.
	PaymentMethod string `json:"payment_method"`
	// Unique identifier to link a payment.
	PaymentContextToken string `json:"payment_context_token"`

	// Specific fields for onchain payments
	Chain string `json:"chain"`
	Asset string `json:"asset"`
}

type PayReq struct {
	// Lightning invoice
	LightningInvoice string `json:"lightning_invoice"`

	// Onchain payment details
	Address string `json:"address"`
	Asset   string `json:"asset"`
	Chain   string `json:"chain"`

	// CheckoutURL is a web based payment flow, used in CC for example.
	CheckoutURL string `json:"check_url"`
}

type PaymentRequestResponse struct {
	// What version the server implements
	Version string `json:"version"`
	// When does this offer expire
	ExpiresAt string `json:"expires_at"`

	// Payment Request details
	PaymentRequest PayReq `json:"payment_request"`
}
