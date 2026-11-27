package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/Vitorpignataro/unified-data-platform/internal/cdc"
)

var (
	ErrInvalidEvent = errors.New("evento CDC inválido")
	ErrTransient    = errors.New("falha transitória")
)

type Message struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
}

type Result struct {
	Topic     string
	Partition int
	Offset    int64

	Database  string
	Table     string
	Operation string
	EntityID  int64

	Tombstone bool
}

type sourceConfig struct {
	Database   string
	Table      string
	PrimaryKey string
}

type Handler interface {
	Handle(ctx context.Context, msg Message) (Result, error)
}

type ValidationHandler struct{}

var supportedTopics = map[string]sourceConfig{
	"udp.cadastro.db_cadastro.empresas": {
		Database:   "db_cadastro",
		Table:      "empresas",
		PrimaryKey: "id_empresa",
	},
	"udp.cadastro.db_cadastro.unidades": {
		Database:   "db_cadastro",
		Table:      "unidades",
		PrimaryKey: "id_unidade",
	},
	"udp.financeiro.db_financeiro.faturas": {
		Database:   "db_financeiro",
		Table:      "faturas",
		PrimaryKey: "id_fatura",
	},
	"udp.financeiro.db_financeiro.pagamentos": {
		Database:   "db_financeiro",
		Table:      "pagamentos",
		PrimaryKey: "id_pagamento",
	},
}

func (ValidationHandler) Handle(
	ctx context.Context,
	msg Message,
) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	source, ok := supportedTopics[msg.Topic]
	if !ok {
		return Result{}, fmt.Errorf(
			"%w: tópico não suportado: %s",
			ErrInvalidEvent,
			msg.Topic,
		)
	}

	if msg.Partition < 0 || msg.Offset < 0 {
		return Result{}, fmt.Errorf(
			"%w: posição Kafka inválida",
			ErrInvalidEvent,
		)
	}

	event, err := cdc.Parse(msg.Value)
	if err != nil {
		return Result{}, fmt.Errorf(
			"%w: %v",
			ErrInvalidEvent,
			err,
		)
	}

	result := Result{
		Topic:     msg.Topic,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Database:  source.Database,
		Table:     source.Table,
	}

	if event.Tombstone {
		if len(msg.Key) == 0 {
			return Result{}, fmt.Errorf(
				"%w: tombstone sem chave",
				ErrInvalidEvent,
			)
		}

		result.Operation = "tombstone"
		result.Tombstone = true

		return result, nil
	}

	// Verifica se a origem corresponde ao tópico.
	if event.Database != source.Database ||
		event.Table != source.Table {
		return Result{}, fmt.Errorf(
			"%w: origem %s.%s não corresponde ao tópico %s",
			ErrInvalidEvent,
			event.Database,
			event.Table,
			msg.Topic,
		)
	}

	var row json.RawMessage

	if event.Operation == "d" {
		row = event.Before
	} else {
		row = event.After
	}

	if len(row) == 0 ||
		bytes.Equal(bytes.TrimSpace(row), []byte("null")) {
		return Result{}, fmt.Errorf(
			"%w: registro ausente para operação %s",
			ErrInvalidEvent,
			event.Operation,
		)
	}

	var fields map[string]json.RawMessage

	if err := json.Unmarshal(row, &fields); err != nil {
		return Result{}, fmt.Errorf(
			"%w: registro malformado: %v",
			ErrInvalidEvent,
			err,
		)
	}

	rawID, exists := fields[source.PrimaryKey]
	if !exists {
		return Result{}, fmt.Errorf(
			"%w: chave %s ausente",
			ErrInvalidEvent,
			source.PrimaryKey,
		)
	}

	id, err := strconv.ParseInt(string(rawID), 10, 64)
	if err != nil || id <= 0 {
		return Result{}, fmt.Errorf(
			"%w: identificador inválido: %s",
			ErrInvalidEvent,
			string(rawID),
		)
	}

	result.Operation = event.Operation
	result.EntityID = id

	return result, nil
}
