package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"

	"github.com/Vitorpignataro/unified-data-platform/internal/cdc"
)

var topics = []string{
	"udp.cadastro.db_cadastro.empresas",
	"udp.cadastro.db_cadastro.unidades",
	"udp.financeiro.db_financeiro.faturas",
	"udp.financeiro.db_financeiro.pagamentos",
}

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	broker := flag.String(
		"broker",
		"localhost:9092",
		"Endereço do broker Kafka",
	)

	group := flag.String(
		"group",
		"udp-data-consumers",
		"Identificador do Consumer Group",
	)

	limit := flag.Int(
		"limit",
		10,
		"Quantidade de eventos (0 = contínuo)",
	)

	flag.Parse()

	if *limit < 0 {
		return fmt.Errorf("limit não pode ser negativo")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{*broker},

		GroupID:     *group,
		GroupTopics: topics,

		GroupBalancers: []kafka.GroupBalancer{
			kafka.RoundRobinGroupBalancer{},
		},

		StartOffset: kafka.FirstOffset,

		MinBytes: 1,
		MaxBytes: 10e6,

		// Commits síncronos quando forem implementados.
		CommitInterval: 0,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Printf("erro ao fechar reader: %v", err)
		}
	}()

	log.Printf("Consumer Group: %s", *group)
	log.Printf("Broker: %s", *broker)
	log.Printf("Tópicos: %v", topics)
	log.Println("Modo diagnóstico: offsets NÃO serão confirmados")

	total := 0

	for *limit == 0 || total < *limit {
		message, err := reader.FetchMessage(ctx)

		if err != nil {
			if errors.Is(err, context.Canceled) &&
				ctx.Err() != nil {
				log.Println("Encerramento solicitado")
				break
			}

			return fmt.Errorf("buscar mensagem Kafka: %w", err)
		}

		event, err := cdc.Parse(message.Value)
		if err != nil {
			return fmt.Errorf(
				"interpretar %s[%d] offset=%d: %w",
				message.Topic,
				message.Partition,
				message.Offset,
				err,
			)
		}

		total++

		// Evita gerar milhares de linhas no terminal.
		if total <= 10 || total%1000 == 0 {
			if event.Tombstone {
				log.Printf(
					"topic=%s partition=%d offset=%d op=TOMBSTONE",
					message.Topic,
					message.Partition,
					message.Offset,
				)
			} else {
				log.Printf(
					"topic=%s partition=%d offset=%d origem=%s.%s op=%s",
					message.Topic,
					message.Partition,
					message.Offset,
					event.Database,
					event.Table,
					event.Operation,
				)
			}
		}

		// NÃO confirmar o offset nesta etapa.
		// O commit só acontecerá depois de persistir
		// o evento com sucesso no PostgreSQL.
	}

	log.Printf("Total de mensagens lidas: %d", total)

	return nil
}
