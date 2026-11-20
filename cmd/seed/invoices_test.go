package main

import "testing"

func TestInvoiceUnitDistribution(t *testing.T) {
	tests := []struct {
		invoiceID      int64
		expectedUnitID int64
	}{
		{1, 1},
		{10000, 10000},
		{10001, 1},
		{20001, 1},
		{100000, 10000},
	}

	for _, tt := range tests {
		invoice := invoiceFor(tt.invoiceID, 10000, 60000)

		if invoice.UnitID != tt.expectedUnitID {
			t.Errorf(
				"fatura %d: unidade esperada %d, recebida %d",
				tt.invoiceID,
				tt.expectedUnitID,
				invoice.UnitID,
			)
		}
	}
}

func TestInvoiceReferenceMonth(t *testing.T) {
	tests := []struct {
		id    int64
		month string
	}{
		{1, "2026-01"},
		{10001, "2026-02"},
		{50001, "2026-06"},
		{100000, "2026-10"},
	}

	for _, tt := range tests {
		invoice := invoiceFor(tt.id, 10000, 60000)

		got := invoice.ReferenceMonth.Format("2006-01")

		if got != tt.month {
			t.Errorf(
				"fatura %d: mês esperado %s, recebido %s",
				tt.id,
				tt.month,
				got,
			)
		}
	}
}

func TestInvoiceStatus(t *testing.T) {
	paid := invoiceFor(60000, 10000, 60000)
	pending := invoiceFor(60001, 10000, 60000)

	if paid.Status != "paga" {
		t.Errorf("status esperado paga, recebido %s", paid.Status)
	}

	if pending.Status != "pendente" {
		t.Errorf("status esperado pendente, recebido %s", pending.Status)
	}
}

func TestInvoiceAmount(t *testing.T) {
	invoice := invoiceFor(42, 10000, 60000)

	if invoice.AmountCents <= 0 {
		t.Fatal("valor da fatura deve ser positivo")
	}

	if formatAmount(15001) != "150.01" {
		t.Fatal("formatação monetária incorreta")
	}
}

func TestInvoiceIsDeterministic(t *testing.T) {
	first := invoiceFor(12345, 10000, 60000)
	second := invoiceFor(12345, 10000, 60000)

	if first != second {
		t.Fatal("geração de faturas não é determinística")
	}
}
