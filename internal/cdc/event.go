package cdc

import (
	"bytes"
	"encoding/json"
	"fmt"
)

type Event struct {
	Operation string
	Database  string
	Table     string

	Before json.RawMessage
	After  json.RawMessage

	Tombstone bool
}

type envelope struct {
	Payload *struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
		Op     string          `json:"op"`
		Source struct {
			DB    string `json:"db"`
			Table string `json:"table"`
		} `json:"source"`
	} `json:"payload"`
}

func Parse(value []byte) (Event, error) {
	raw := bytes.TrimSpace(value)

	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return Event{Tombstone: true}, nil
	}

	var data envelope

	if err := json.Unmarshal(raw, &data); err != nil {
		return Event{}, fmt.Errorf(
			"decodificar evento Debezium: %w",
			err,
		)
	}

	if data.Payload == nil {
		return Event{}, fmt.Errorf(
			"evento Debezium sem payload",
		)
	}

	p := data.Payload

	switch p.Op {
	case "r", "c", "u", "d":
	default:
		return Event{}, fmt.Errorf(
			"operação não suportada: %q",
			p.Op,
		)
	}

	if p.Source.DB == "" || p.Source.Table == "" {
		return Event{}, fmt.Errorf(
			"evento sem identificação da origem",
		)
	}

	return Event{
		Operation: p.Op,
		Database:  p.Source.DB,
		Table:     p.Source.Table,
		Before:    p.Before,
		After:     p.After,
	}, nil
}
