package main

import "testing"

func TestCompanyFor(t *testing.T) {
	company := companyFor(1)

	if company.ID != 1 {
		t.Errorf("ID esperado 1, recebido %d", company.ID)
	}

	if company.Name != "Empresa Fictícia 0001" {
		t.Errorf("nome inesperado: %s", company.Name)
	}

	if len(company.CNPJ) != 14 {
		t.Errorf(
			"CNPJ deve ter 14 caracteres, recebeu %d",
			len(company.CNPJ),
		)
	}
}

func TestCompanyForIsDeterministic(t *testing.T) {
	first := companyFor(42)
	second := companyFor(42)

	if first != second {
		t.Fatal("a geração deve ser determinística")
	}
}
