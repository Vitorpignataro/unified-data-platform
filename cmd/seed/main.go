package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Source struct {
	Name   string
	Prefix string
	Port   string
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cadastro := Source{
		Name:   "Cadastro",
		Prefix: "MYSQL_A",
		Port:   "3307",
	}

	financeiro := Source{
		Name:   "Financeiro",
		Prefix: "MYSQL_B",
		Port:   "3308",
	}

	dbCadastro, err := openSource(cadastro)
	if err != nil {
		return err
	}
	defer dbCadastro.Close()

	dbFinanceiro, err := openSource(financeiro)
	if err != nil {
		return err
	}
	defer dbFinanceiro.Close()

	fmt.Println("Conexões estabelecidas!")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		2*time.Minute,
	)
	defer cancel()

	fmt.Println("Gerando empresas...")

	if err := GenerateCompanies(ctx, dbCadastro, 1000); err != nil {
		return err
	}

	var total int

	err = dbCadastro.QueryRowContext(
		ctx,
		"SELECT COUNT(*) FROM empresas",
	).Scan(&total)

	if err != nil {
		return fmt.Errorf("contar empresas: %w", err)
	}

	fmt.Printf("Empresas no banco: %d\n", total)

	return nil
}

func openSource(source Source) (*sql.DB, error) {
	getenv := func(key string) string {
		return os.Getenv(source.Prefix + "_" + key)
	}

	if getenv("USER") == "" ||
		getenv("PASSWORD") == "" ||
		getenv("DB") == "" {
		return nil, fmt.Errorf(
			"%s: variáveis de ambiente incompletas",
			source.Name,
		)
	}

	cfg := mysql.NewConfig()
	cfg.User = getenv("USER")
	cfg.Passwd = getenv("PASSWORD")
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort("127.0.0.1", source.Port)
	cfg.DBName = getenv("DB")
	cfg.ParseTime = true

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf(
			"conexão %s: %w",
			source.Name,
			err,
		)
	}

	fmt.Printf("[%s] Conectado!\n", source.Name)

	return db, nil
}
