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
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/Vitorpignataro/unified-data-platform/internal/consumer"
)

// Tópicos monitorados pelo CDC Consumer.
var topics = []string{
	"udp.cadastro.db_cadastro.empresas",
	"udp.cadastro.db_cadastro.unidades",
	"udp.financeiro.db_financeiro.faturas",
	"udp.financeiro.db_financeiro.pagamentos",
}

func main() {
	// Permite encerramento gracioso com Ctrl+C ou SIGTERM.
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
	// Configurações via argumentos.
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
		"Quantidade de mensagens (0 = contínuo)",
	)

	flag.Parse()

	if *limit < 0 {
		return fmt.Errorf("limit não pode ser negativo")
	}

	// Configuração do Kafka Reader.
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
			log.Printf(
				"erro ao fechar Kafka Reader: %v",
				err,
			)
		}
	}()

	// Handler responsável por validar eventos CDC.
	handler := consumer.ValidationHandler{}

	log.Println("====================================")
	log.Println(" Unified Data Platform - CDC Consumer")
	log.Println("====================================")
	log.Printf("Broker: %s", *broker)
	log.Printf("Consumer Group: %s", *group)
	log.Printf("Tópicos monitorados: %d", len(topics))
	log.Println("Modo diagnóstico: offsets NÃO serão confirmados")
	log.Println("Aguardando mensagens...")

	total := 0

	// Limit = 0 mantém o consumer em execução contínua.
	for *limit == 0 || total < *limit {

		// Busca mensagem sem confirmar o offset.
		message, err := reader.FetchMessage(ctx)

		if err != nil {
			if errors.Is(err, context.Canceled) &&
				ctx.Err() != nil {
				log.Println("Encerramento solicitado")
				break
			}

			if ctx.Err() != nil {
				log.Println("Contexto encerrado")
				break
			}

			return fmt.Errorf(
				"erro ao buscar mensagem Kafka: %w",
				err,
			)
		}

		// Converte a mensagem Kafka para nosso modelo interno.
		msg := consumer.Message{
			Topic:     message.Topic,
			Partition: message.Partition,
			Offset:    message.Offset,
			Key:       message.Key,
			Value:     message.Value,
		}

		// Processa a mensagem com validação e retry.
		result, err := consumer.ProcessWithRetry(
			ctx,
			handler,
			msg,
			3,
			200*time.Millisecond,
		)

		if err != nil {
			// Encerramento durante o processamento.
			if ctx.Err() != nil {
				log.Println(
					"Processamento interrompido pelo encerramento",
				)
				break
			}

			// Diferencia falhas de validação.
			if errors.Is(err, consumer.ErrInvalidEvent) {
				return fmt.Errorf(
					"evento inválido topic=%s partition=%d offset=%d: %w",
					message.Topic,
					message.Partition,
					message.Offset,
					err,
				)
			}

			return fmt.Errorf(
				"falha no processamento topic=%s partition=%d offset=%d: %w",
				message.Topic,
				message.Partition,
				message.Offset,
				err,
			)
		}

		total++

		// Evita excesso de logs durante grandes snapshots.
		if total <= 10 || total%1000 == 0 {

			if result.Tombstone {
				log.Printf(
					"TOMBSTONE topic=%s partition=%d offset=%d",
					result.Topic,
					result.Partition,
					result.Offset,
				)
			} else {
				log.Printf(
					"VALIDADO topic=%s partition=%d offset=%d origem=%s.%s op=%s id=%d",
					result.Topic,
					result.Partition,
					result.Offset,
					result.Database,
					result.Table,
					result.Operation,
					result.EntityID,
				)
			}
		}

		// IMPORTANTE:
		//
		// Ainda não utilizamos CommitMessages().
		//
		// Nesta fase, apenas validamos eventos.
		// A confirmação dos offsets será implementada
		// após a persistência no PostgreSQL.
	}

	log.Println("====================================")
	log.Printf("Total de mensagens validadas: %d", total)
	log.Println("CDC Consumer finalizado")
	log.Println("====================================")

	return nil
}
