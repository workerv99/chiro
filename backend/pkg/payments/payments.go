package payments

import (
	"fmt"
	"os"
)

var (
	stripeKey      string
	stripePriceID  string
	paypalClientID string
	paypalSecret   string
)

func Init() {
	stripeKey = os.Getenv("STRIPE_SECRET_KEY")
	stripePriceID = os.Getenv("STRIPE_PRICE_ID")
	paypalClientID = os.Getenv("PAYPAL_CLIENT_ID")
	paypalSecret = os.Getenv("PAYPAL_SECRET")
}

func IsStripeConfigured() bool {
	return stripeKey != "" && stripePriceID != ""
}

func IsPayPalConfigured() bool {
	return paypalClientID != "" && paypalSecret != ""
}

type CheckoutSession struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	Provider  string `json:"provider"`
}

func CreateStripeCheckout(userID, email, successURL, cancelURL string) (*CheckoutSession, error) {
	if !IsStripeConfigured() {
		return nil, fmt.Errorf("Stripe no configurado")
	}

	// Placeholder - integrar con stripe-go SDK
	// session, err := session.New(&stripe.CheckoutSessionParams{
	//   Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
	//   CustomerEmail: stripe.String(email),
	//   LineItems: []*stripe.CheckoutSessionLineItemParams{
	//     {Price: stripe.String(stripePriceID), Quantity: stripe.Int64(1)},
	//   },
	//   SuccessURL: stripe.String(successURL),
	//   CancelURL:  stripe.String(cancelURL),
	//   Metadata:   map[string]string{"user_id": userID},
	// })

	return &CheckoutSession{
		SessionID: "placeholder",
		URL:       successURL,
		Provider:  "stripe",
	}, nil
}
