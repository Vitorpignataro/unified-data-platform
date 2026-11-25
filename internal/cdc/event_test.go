package cdc

import "testing"

func TestParseSnapshot(t *testing.T) {
	value := []byte(`{
		"schema": {},
		"payload": {
			"before": null,
			"after": {
				"id_empresa": 1,
				"razao_social": "Empresa Fictícia 0001"
			},
			"source": {
				"db": "db_cadastro",
				"table": "empresas"
			},
			"op": "r"
		}
	}`)

	event, err := Parse(value)
	if err != nil {
		t.Fatal(err)
	}

	if event.Operation != "r" {
		t.Errorf("operação incorreta: %s", event.Operation)
	}

	if event.Database != "db_cadastro" {
		t.Errorf("banco incorreto: %s", event.Database)
	}

	if event.Table != "empresas" {
		t.Errorf("tabela incorreta: %s", event.Table)
	}

	if event.Tombstone {
		t.Fatal("snapshot não deve ser tombstone")
	}
}

func TestParseTombstone(t *testing.T) {
	event, err := Parse(nil)
	if err != nil {
		t.Fatal(err)
	}

	if !event.Tombstone {
		t.Fatal("tombstone não identificado")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	_, err := Parse([]byte(`{invalid`))

	if err == nil {
		t.Fatal("JSON inválido deveria retornar erro")
	}
}

func TestParseUnknownOperation(t *testing.T) {
	value := []byte(`{
		"payload": {
			"op": "x",
			"source": {
				"db": "db_cadastro",
				"table": "empresas"
			}
		}
	}`)

	_, err := Parse(value)

	if err == nil {
		t.Fatal("operação desconhecida deveria falhar")
	}
}
