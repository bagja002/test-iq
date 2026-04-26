package services

import (
	"testing"

	"test-iq-ku/apps/api/internal/models"
)

func TestMapMidtransTransactionStatus(t *testing.T) {
	testCases := []struct {
		name              string
		transactionStatus string
		fraudStatus       string
		expected          models.PaymentStatus
	}{
		{name: "settlement paid", transactionStatus: "settlement", expected: models.PaymentStatusPaid},
		{name: "capture accept paid", transactionStatus: "capture", fraudStatus: "accept", expected: models.PaymentStatusPaid},
		{name: "pending pending", transactionStatus: "pending", expected: models.PaymentStatusPending},
		{name: "expire expired", transactionStatus: "expire", expected: models.PaymentStatusExpired},
		{name: "cancel canceled", transactionStatus: "cancel", expected: models.PaymentStatusCanceled},
		{name: "deny failed", transactionStatus: "deny", expected: models.PaymentStatusFailed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if actual := mapMidtransTransactionStatus(tc.transactionStatus, tc.fraudStatus); actual != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, actual)
			}
		})
	}
}

func TestBuildMidtransSignature(t *testing.T) {
	signature := buildMidtransSignature("ORDER-1", "200", "50000.00", "server-key")
	expected := "20be7ffce397fbc540934b137d0c27901461f8587d82dbaff2fa298c8438c685740e3110f76cd0822e87241de3244c0c61cca218d9ab164e489085a9e66b5f66"
	if signature != expected {
		t.Fatalf("expected signature %s, got %s", expected, signature)
	}
}
