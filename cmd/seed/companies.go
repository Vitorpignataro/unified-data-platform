package main

import (
	"context"
	"database/sql"
	"fmt"
)

type Company struct {
	ID   int64
	Name string
	CNPJ string
}

func companyFor(id int64) Company {
	return Company{
		ID:   id,
		Name: fmt.Sprintf("Empresa Fictícia %04d", id),
		CNPJ: fmt.Sprintf("99%012d", id),
	}
}

func GenerateCompanies(
	ctx context.Context,
	db *sql.DB,
	total int,
) error {
	if total <= 0 {
		return fmt.Errorf("total de empresas deve ser positivo")
	}

	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO empresas (
			id_empresa,
			razao_social,
			cnpj
		)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id_empresa = id_empresa
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparar insert: %w", err)
	}
	defer stmt.Close()

	for i := 1; i <= total; i++ {
		company := companyFor(int64(i))

		_, err := stmt.ExecContext(
			ctx,
			company.ID,
			company.Name,
			company.CNPJ,
		)

		if err != nil {
			return fmt.Errorf(
				"inserir empresa %d: %w",
				company.ID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar transação: %w", err)
	}

	return nil
}
