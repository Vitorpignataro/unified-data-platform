package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Vitorpignataro/unified-data-platform/internal/consumer"
)

func TestSaveEventIntegration(t *testing.T) {
	if os.Getenv("RUN_PG_INTEGRATION") != "1" {
		t.Skip("teste de integração não habilitado")
	}

	cfg, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer cancel()

	repo, err := New(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Tópico exclusivo para este teste.
	topic := fmt.Sprintf(
		"udp.test.repository.%d",
		time.Now().UnixNano(),
	)

	// Remove os dados de teste ao finalizar.
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		_, err := repo.pool.Exec(
			cleanupCtx,
			"DELETE FROM raw.kafka_messages WHERE topic = $1",
			topic,
		)
		if err != nil {
			t.Errorf("limpar teste: %v", err)
		}

		repo.Close()
	})

	msg := consumer.Message{
		Topic:     topic,
		Partition: 0,
		Offset:    1,
		Key:       []byte(`{"payload":{"id_empresa":1}}`),
		Value:     []byte(`{"payload":{"op":"c"}}`),
	}

	result := consumer.Result{
		Topic:     topic,
		Partition: 0,
		Offset:    1,
		Database:  "db_cadastro",
		Table:     "empresas",
		Operation: "c",
		EntityID:  1,
	}

	// Primeira inserção.
	inserted, err := repo.SaveEvent(ctx, msg, result)
	if err != nil {
		t.Fatal(err)
	}

	if !inserted {
		t.Fatal("primeiro evento deveria ser inserido")
	}

	// Segunda inserção da mesma posição Kafka.
	inserted, err = repo.SaveEvent(ctx, msg, result)
	if err != nil {
		t.Fatal(err)
	}

	if inserted {
		t.Fatal("evento duplicado não deveria ser inserido")
	}

	var total int

	err = repo.pool.QueryRow(
		ctx,
		`SELECT COUNT(*)
		 FROM raw.kafka_messages
		 WHERE topic = $1`,
		topic,
	).Scan(&total)

	if err != nil {
		t.Fatal(err)
	}

	if total != 1 {
		t.Fatalf("esperado 1 evento, recebido %d", total)
	}

	// Teste de tombstone.
	tombstone := consumer.Message{
		Topic:     topic,
		Partition: 0,
		Offset:    2,
		Key:       []byte(`{"payload":{"id_empresa":1}}`),
		Value:     nil,
	}

	tombResult := consumer.Result{
		Topic:     topic,
		Partition: 0,
		Offset:    2,
		Database:  "db_cadastro",
		Table:     "empresas",
		Operation: "tombstone",
		Tombstone: true,
	}

	inserted, err = repo.SaveEvent(
		ctx,
		tombstone,
		tombResult,
	)
	if err != nil {
		t.Fatal(err)
	}

	if !inserted {
		t.Fatal("tombstone deveria ser inserido")
	}

	var valueIsNull, entityIsNull bool

	err = repo.pool.QueryRow(
		ctx,
		`SELECT
		    message_value IS NULL,
		    entity_id IS NULL
		 FROM raw.kafka_messages
		 WHERE topic = $1 AND offset_id = 2`,
		topic,
	).Scan(&valueIsNull, &entityIsNull)

	if err != nil {
		t.Fatal(err)
	}

	if !valueIsNull || !entityIsNull {
		t.Fatal("tombstone armazenado incorretamente")
	}
}
