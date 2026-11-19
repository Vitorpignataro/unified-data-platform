package main

import (
	"context"
	"database/sql"
	"fmt"
)

type Unit struct {
	ID        int64
	CompanyID int64
	Name      string
	City      string
	UF        string
	Status    string
}

type Location struct {
	City string
	UF   string
}

var locations = []Location{
	{"São Paulo", "SP"},
	{"Sorocaba", "SP"},
	{"Campinas", "SP"},
	{"Curitiba", "PR"},
	{"Rio de Janeiro", "RJ"},
	{"Belo Horizonte", "MG"},
	{"Florianópolis", "SC"},
	{"Porto Alegre", "RS"},
}

func unitFor(id int64, companyTotal int64) Unit {
	companyID := ((id - 1) % companyTotal) + 1

	locationIndex := (id - 1 + (id-1)/companyTotal) % int64(len(locations))

	location := locations[locationIndex]

	status := "ativa"

	if id%10 == 0 {
		status = "inativa"
	}

	return Unit{
		ID:        id,
		CompanyID: companyID,
		Name:      fmt.Sprintf("Unidade Fictícia %05d", id),
		City:      location.City,
		UF:        location.UF,
		Status:    status,
	}
}

func GenerateUnits(
	ctx context.Context,
	db *sql.DB,
	total int,
	companyTotal int,
) error {
	if total <= 0 || companyTotal <= 0 {
		return fmt.Errorf("quantidades devem ser positivas")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciar transação: %w", err)
	}

	defer tx.Rollback()

	query := `
		INSERT INTO unidades (
			id_unidade,
			id_empresa,
			nome,
			cidade,
			uf,
			status
		)
		VALUES (?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			id_unidade = id_unidade
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("preparar insert unidades: %w", err)
	}
	defer stmt.Close()

	for i := 1; i <= total; i++ {
		unit := unitFor(int64(i), int64(companyTotal))

		_, err := stmt.ExecContext(
			ctx,
			unit.ID,
			unit.CompanyID,
			unit.Name,
			unit.City,
			unit.UF,
			unit.Status,
		)

		if err != nil {
			return fmt.Errorf(
				"inserir unidade %d: %w",
				unit.ID,
				err,
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmar unidades: %w", err)
	}
	fmt.Printf("%d unidades processadas\n", total)
	return nil
}
