package main

import "testing"

func TestUnitFor(t *testing.T) {
	tests := []struct {
		unitID            int64
		expectedCompanyID int64
	}{
		{1, 1},
		{1000, 1000},
		{1001, 1},
		{2001, 1},
		{10000, 1000},
	}

	for _, tt := range tests {
		unit := unitFor(tt.unitID, 1000)

		if unit.CompanyID != tt.expectedCompanyID {
			t.Errorf(
				"unidade %d: esperado empresa %d, recebido %d",
				tt.unitID,
				tt.expectedCompanyID,
				unit.CompanyID,
			)
		}
	}
}

func TestUnitForIsDeterministic(t *testing.T) {
	first := unitFor(1500, 1000)
	second := unitFor(1500, 1000)

	if first != second {
		t.Fatal("geração não é determinística")
	}
}

func TestUnitStatus(t *testing.T) {
	active := unitFor(1, 1000)
	inactive := unitFor(10, 1000)

	if active.Status != "ativa" {
		t.Errorf("status inesperado: %s", active.Status)
	}

	if inactive.Status != "inativa" {
		t.Errorf("status inesperado: %s", inactive.Status)
	}
}

func TestUnitLocation(t *testing.T) {
	unit := unitFor(123, 1000)

	if unit.City == "" || unit.UF == "" {
		t.Fatal("cidade e UF não podem estar vazios")
	}
}
