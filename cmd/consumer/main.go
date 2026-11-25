package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Vitorpignataro/unified-data-platform/internal/cdc"
	"github.com/segmentio/kafka-go"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	broker := flag.String(
		"broker",
		"localhost:9092",
		"Endereço do broker Kafka",
	)

	topic := flag.String(
		"topic",
		"udp.cadastro.db_cadastro_empresas",
		"Tópica Kafka",
	)

	limit := flag.Int(
		"limit",
		5,
		"Quantidade de mensagens a ler",
	)

	flag.Parse()

	if *limit <= 0 {
		return fmt.Errorf("limit deve ser positivo")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{*broker},
		Topic:       *topic,
		Partition:   0,
		StartOffset: kafka.FirstOffset,
		MinBytes:    1,
		MaxBytes:    10e6,
	})

	defer reader.Close()

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)

	defer cancel()

	fmt.Printf("Conectado ao tópico: %s\n", *topic)

	for i := 0; i < *limit; i++ {
		message, err := reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf(
				"ler mensagem Kafka: %w",
				err,
			)
		}

		event, err := cdc.Parse(message.Value)
		if err != nil {
			return fmt.Errorf(
				"interpretar offset %d: %w",
				message.Offset,
				err,
			)
		}

		if event.Tombstone {
			fmt.Printf(
				"offset=%d operação=TOMBSTONE\n",
				message.Offset,
			)
			continue
		}

		fmt.Printf(
			"topic=%s offset=%d origem=%s.%s op=%s\n",
			message.Topic,
			message.Offset,
			event.Database,
			event.Table,
			event.Operation,
		)
	}
	return nil
}
