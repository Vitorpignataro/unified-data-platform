package main

import (
	"testing"
	"time"
)

func TestPaymentFor(t *testing.T) {
	invoice := PayableInvoice{
		ID:             1,
		Amount:         "1250.75",
		ReferenceMonth: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	payment := paymentFor(invoice)

	if payment.ID != invoice.ID {
		t.Errorf("ID do pagamento incorreto: %d", payment.ID)
	}

	if payment.InvoiceID != invoice.ID {
		t.Errorf("ID da fatura incorreto: %d", payment.InvoiceID)
	}

	if payment.Amount != invoice.Amount {
		t.Errorf(
			"valor esperado %s, recebido %s",
			invoice.Amount,
			payment.Amount,
		)
	}
}

func TestPaymentDate(t *testing.T) {
	tests := []struct {
		invoiceID int64
		month     time.Month
		expected  string
	}{
		{1, time.January, "2026-02-06"},
		{10, time.January, "2026-02-05"},
		{60000, time.June, "2026-07-05"},
	}

	for _, tt := range tests {
		invoice := PayableInvoice{
			ID:     tt.invoiceID,
			Amount: "500.00",
			ReferenceMonth: time.Date(
				2026, tt.month, 1,
				0, 0, 0, 0, time.UTC,
			),
		}

		payment := paymentFor(invoice)
		got := payment.PaymentDate.Format("2006-01-02")

		if got != tt.expected {
			t.Errorf(
				"fatura %d: esperado %s, recebido %s",
				tt.invoiceID,
				tt.expected,
				got,
			)
		}
	}
}

func TestPaymentIsDeterministic(t *testing.T) {
	invoice := PayableInvoice{
		ID:     500,
		Amount: "1599.99",
		ReferenceMonth: time.Date(
			2026, time.March, 1,
			0, 0, 0, 0, time.UTC,
		),
	}

	first := paymentFor(invoice)
	second := paymentFor(invoice)

	if first != second {
		t.Fatal("geração de pagamento não é determinística")
	}
}
