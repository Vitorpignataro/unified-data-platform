package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

const paymentBatchSize = 1000

type PayableInvoice struct {
	ID             int64
	Amount         string
	ReferenceMonth time.Time
}

type Payment struct {
	ID          int64
	InvoiceID   int64
	Amount      string
	PaymentDate time.Time
}

func paymentFor(invoice PayableInvoice) Payment {
	dayOffset := 4 + int(invoice.ID%10)

	paymentDate := invoice.ReferenceMonth.AddDate(
		0,
		1,
		dayOffset,
	)

	return Payment{
		ID:          invoice.ID,
		InvoiceID:   invoice.ID,
		Amount:      invoice.Amount,
		PaymentDate: paymentDate,
	}
}

func loadPaidInvoices(
	ctx context.Context,
	db *sql.DB,
	afterID int64,
) ([]PayableInvoice, error) {
	query := `
		SELECT
			id_fatura,
			valor_total,
			mes_referencia
		FROM faturas
		WHERE status = 'paga'
		  AND id_fatura > ?
		ORDER BY id_fatura
		LIMIT ?
	`

	rows, err := db.QueryContext(
		ctx,
		query,
		afterID,
		paymentBatchSize,
	)

	if err != nil {
		return nil, fmt.Errorf("consultar faturas: %w", err)
	}
	defer rows.Close()

	invoices := make([]PayableInvoice, 0, paymentBatchSize)

	for rows.Next() {
		var invoice PayableInvoice

		if err := rows.Scan(
			&invoice.ID,
			&invoice.Amount,
			&invoice.ReferenceMonth,
		); err != nil {
			return nil, fmt.Errorf("ler fatura: %w", err)
		}
		invoices = append(invoices, invoice)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterar faturas: %w", err)
	}

	return invoices, nil

}

func insertPaymentBatch(
	ctx context.Context,
	db *sql.DB,
	invoices []PayableInvoice,
) error {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pagamentos (
			id_pagamento,
			id_fatura,
			valor_pago,
			data_pagamento
		)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id_pagamento = id_pagamento
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparar insert: %w", err)
	}
	defer stmt.Close()

	for _, invoice := range invoices {
		payment := paymentFor(invoice)

		_, err := stmt.ExecContext(
			ctx,
			payment.ID,
			payment.InvoiceID,
			payment.Amount,
			payment.PaymentDate.Format("2006-01-02"),
		)

		if err != nil {
			return fmt.Errorf(
				"inserir pagamento %d: %w",
				payment.ID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar lote: %w", err)
	}

	return nil

}

func GeneratePayments(
	ctx context.Context,
	db *sql.DB,
	expectedPaid int,
) error {
	if expectedPaid <= 0 {
		return fmt.Errorf("total esperado deve ser positivo")
	}

	var paidInvoices int

	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		 FROM faturas
		 WHERE status = 'paga'`,
	).Scan(&paidInvoices)

	if err != nil {
		return fmt.Errorf("contar faturas pagas: %w", err)
	}

	if paidInvoices != expectedPaid {
		return fmt.Errorf(
			"esperadas %d faturas pagas, encontradas %d",
			expectedPaid,
			paidInvoices,
		)
	}

	var lastID int64
	totalProcessed := 0

	for {
		invoices, err := loadPaidInvoices(
			ctx,
			db,
			lastID,
		)
		if err != nil {
			return err
		}

		if len(invoices) == 0 {
			break
		}

		if err := insertPaymentBatch(
			ctx,
			db,
			invoices,
		); err != nil {
			return fmt.Errorf(
				"inserir pagamentos após fatura %d: %w",
				lastID,
				err,
			)
		}

		lastID = invoices[len(invoices)-1].ID
		totalProcessed += len(invoices)

		if totalProcessed%10000 == 0 {
			fmt.Printf(
				"Pagamentos processados: %d/%d\n",
				totalProcessed,
				expectedPaid,
			)
		}
	}

	if totalProcessed != expectedPaid {
		return fmt.Errorf(
			"processados %d pagamentos, esperados %d",
			totalProcessed,
			expectedPaid,
		)
	}

	fmt.Printf(
		"Geração concluída: %d pagamentos processados\n",
		totalProcessed,
	)

	return nil
}
