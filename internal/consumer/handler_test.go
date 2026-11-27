package consumer

import (
	"context"
	"errors"
	"testing"
)

func TestValidationHandler(t *testing.T) {
	handler := ValidationHandler{}

	msg := Message{
		Topic:     "udp.financeiro.db_financeiro.faturas",
		Partition: 0,
		Offset:    42,
		Value: []byte(`{
			"payload": {
				"before": null,
				"after": {
					"id_fatura": 100,
					"valor_total": "1250.75"
				},
				"op": "c",
				"source": {
					"db": "db_financeiro",
					"table": "faturas"
				}
			}
		}`),
	}

	result, err := handler.Handle(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}

	if result.Operation != "c" {
		t.Errorf("operação inesperada: %s", result.Operation)
	}

	if result.EntityID != 100 {
		t.Errorf("ID esperado 100, recebido %d", result.EntityID)
	}

	if result.Offset != 42 {
		t.Errorf("offset incorreto: %d", result.Offset)
	}
}

func TestValidationHandlerDelete(t *testing.T) {
	handler := ValidationHandler{}

	msg := Message{
		Topic: "udp.cadastro.db_cadastro.empresas",
		Value: []byte(`{
			"payload": {
				"before": {"id_empresa": 50},
				"after": null,
				"op": "d",
				"source": {
					"db": "db_cadastro",
					"table": "empresas"
				}
			}
		}`),
	}

	result, err := handler.Handle(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}

	if result.Operation != "d" || result.EntityID != 50 {
		t.Fatal("DELETE não foi identificado corretamente")
	}
}

func TestValidationHandlerTombstone(t *testing.T) {
	handler := ValidationHandler{}

	msg := Message{
		Topic: "udp.cadastro.db_cadastro.empresas",
		Key:   []byte(`{"payload":{"id_empresa":50}}`),
		Value: nil,
	}

	result, err := handler.Handle(context.Background(), msg)
	if err != nil {
		t.Fatal(err)
	}

	if !result.Tombstone {
		t.Fatal("tombstone não identificado")
	}
}

func TestValidationHandlerInvalidTopic(t *testing.T) {
	handler := ValidationHandler{}

	_, err := handler.Handle(
		context.Background(),
		Message{Topic: "topico.desconhecido"},
	)

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("esperado ErrInvalidEvent, recebido %v", err)
	}
}

func TestValidationHandlerMissingID(t *testing.T) {
	handler := ValidationHandler{}

	msg := Message{
		Topic: "udp.cadastro.db_cadastro.empresas",
		Value: []byte(`{
			"payload": {
				"after": {"razao_social": "Teste"},
				"op": "c",
				"source": {
					"db": "db_cadastro",
					"table": "empresas"
				}
			}
		}`),
	}

	_, err := handler.Handle(context.Background(), msg)

	if !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("esperado ErrInvalidEvent, recebido %v", err)
	}
}
