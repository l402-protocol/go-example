# L402 Setup Example

This repository demonstrates a complete implementation of the L402 protocol, which enables HTTP 402 Payment Required responses with structured payment information. The example consists of three main components that work together to showcase a paywalled service implementation.

## Components

### 1. Client
The client component demonstrates how to consume an application that implements L402 protocol paywall protection. It includes:
- Automatic handling of 402 Payment Required responses
- Integration with a wallet interface for processing payments
- Support for both mock and simulated payment flows

### 2. Server
The server provides a protected service that requires payment for access. Key features:
- Protected endpoints that require authentication
- Integration with the L402 protocol for payment requirements
- Automatic generation of payment challenges (402 responses)
- Verification of completed payments

### 3. Gateway
The gateway handles all payment-related operations, serving as a payment processor. It provides:
- Payment request handling
- Checkout flow management
- Payment verification
- Server notification of successful payments
- Support for multiple payment methods (simulated)

## Getting Started

### Prerequisites
- Go 1.21 or later

### Running the Example

1. Start both the server and gateway:
```bash
make serve-server
make serve-gateway
```

This will start:
- Server on :8080
- Gateway on :8081

2. Run the client:

To see available offers without making a payment:
```bash
go run cmd/client/main.go
```


To simulate a payment for a specific offer:
```bash
go run cmd/client/main.go --fake --offer-id=offer_0001
```

## Flow Description

1. The client attempts to access a protected resource on the server
2. If payment is required, the server returns a 402 response with available offers
3. The client's wallet processes the payment through the gateway
4. The gateway verifies the payment and notifies the server
5. The client can then access the protected resource

## Development

Individual components can be run separately for development:

```bash
make serve-server   # Run only the server
make serve-gateway  # Run only the gateway
```

Clean built binaries:
```bash
make clean
```

