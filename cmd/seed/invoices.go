package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Invoice struct {
	ID             int64
	UnitID         int64
	ReferenceMonth time.Time
	AmountCents    int64
	Status         string
}

func invoiceFor(
	id int64,
	unitTotal int64,
	paidTotal int64,
) Invoice {
	unitID := ((id - 1) % unitTotal) + 1

	monthIndex := (id - 1) / unitTotal

	referenceMonth := time.Date(
		2026,
		time.January,
		1, 0, 0, 0, 0,
		time.UTC,
	).AddDate(0, int(monthIndex), 0)

	amountCents := int64(15000) + (id*137)%500000

	status := "pendente"
	if id <= paidTotal {
		status = "paga"
	}

	return Invoice{
		ID:             id,
		UnitID:         unitID,
		ReferenceMonth: referenceMonth,
		AmountCents:    amountCents,
		Status:         status,
	}

}

func formatAmount(cents int64) string {
	return fmt.Sprintf(
		"%d.%02d",
		cents/100,
		cents%100,
	)
}

const invoiceBatchSize = 1000

func insertInvoiceBatch(
	ctx context.Context,
	db *sql.DB,
	start int,
	end int,
	unitTotal int,
	paidTotal int,
) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO faturas (
			id_fatura,
			id_unidade,
			mes_referencia,
			valor_total,
			status
		)
		VALUES (?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id_fatura = id_fatura
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparar insert: %w", err)
	}
	defer stmt.Close()

	for i := start; i <= end; i++ {
		invoice := invoiceFor(
			int64(i),
			int64(unitTotal),
			int64(paidTotal),
		)

		_, err := stmt.ExecContext(
			ctx,
			invoice.ID,
			invoice.UnitID,
			invoice.ReferenceMonth.Format("2006-01-02"),
			formatAmount(invoice.AmountCents),
			invoice.Status,
		)

		if err != nil {
			return fmt.Errorf(
				"inserir fatura %d: %w",
				invoice.ID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar lote: %w", err)
	}

	return nil
}

func GenerateInvoices(
	ctx context.Context,
	dbCadastro *sql.DB,
	dbFinanceiro *sql.DB,
	total int,
	unitTotal int,
	paidTotal int,
) error {
	if total <= 0 || unitTotal <= 0 {
		return fmt.Errorf("quantidades devem ser positivas")
	}

	if paidTotal < 0 || paidTotal > total {
		return fmt.Errorf("quantidade de faturas pagas inválida")
	}

	var existingUnits int

	err := dbCadastro.QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		 FROM unidades
		 WHERE id_unidade BETWEEN 1 AND ?`,
		unitTotal,
	).Scan(&existingUnits)

	if err != nil {
		return fmt.Errorf("consultar unidades: %w", err)
	}

	if existingUnits != unitTotal {
		return fmt.Errorf(
			"esperadas %d unidades, encontradas %d",
			unitTotal,
			existingUnits,
		)
	}

	for start := 1; start <= total; start += invoiceBatchSize {
		end := start + invoiceBatchSize - 1

		if end > total {
			end = total
		}

		err := insertInvoiceBatch(
			ctx,
			dbFinanceiro,
			start,
			end,
			unitTotal,
			paidTotal,
		)

		if err != nil {
			return fmt.Errorf(
				"lote %d-%d: %w",
				start,
				end,
				err,
			)
		}

		if end%10000 == 0 || end == total {
			fmt.Printf(
				"Faturas processadas: %d/%d\n",
				end,
				total,
			)
		}
	}

	return nil
}
