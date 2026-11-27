package postgres

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/Vitorpignataro/unified-data-platform/internal/consumer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg Config) (*Repository, error) {
	address := net.JoinHostPort(
		cfg.Host,
		strconv.Itoa(cfg.Port),
	)

	connURL := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   address,
		Path:   "/" + cfg.Database,
	}

	query := connURL.Query()
	query.Set("sslmode", "disable")
	connURL.RawQuery = query.Encode()

	poolConfig, err := pgxpool.ParseConfig(connURL.String())
	if err != nil {
		return nil, fmt.Errorf("configurar pool: %w", err)
	}

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 0

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("criar pool PostgreSQL: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("conectar PostgreSQL: %w", err)
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) SaveEvent(
	ctx context.Context,
	msg consumer.Message,
	result consumer.Result,
) (bool, error) {
	// Garante que os metadados pertencem à mensagem.
	if msg.Topic != result.Topic ||
		msg.Partition != result.Partition ||
		msg.Offset != result.Offset {
		return false, fmt.Errorf(
			"posição Kafka inconsistente",
		)
	}

	if msg.Topic == "" ||
		msg.Partition < 0 ||
		msg.Offset < 0 {
		return false, fmt.Errorf("posição Kafka inválida")
	}

	if result.Database == "" || result.Table == "" {
		return false, fmt.Errorf("origem CDC inválida")
	}

	var entityID *int64

	if result.Tombstone {
		if result.Operation != "tombstone" ||
			len(msg.Key) == 0 ||
			len(msg.Value) != 0 {
			return false, fmt.Errorf("tombstone inválido")
		}
	} else {
		switch result.Operation {
		case "r", "c", "u", "d":
		default:
			return false, fmt.Errorf(
				"operação inválida: %s",
				result.Operation,
			)
		}

		if len(msg.Value) == 0 || result.EntityID <= 0 {
			return false, fmt.Errorf("evento CDC incompleto")
		}

		entityID = &result.EntityID
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("iniciar transação: %w", err)
	}
	defer tx.Rollback(context.Background())

	query := `
		INSERT INTO raw.kafka_messages (
			topic,
			partition_id,
			offset_id,
			message_key,
			message_value,
			source_database,
			source_table,
			operation,
			entity_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (
			topic,
			partition_id,
			offset_id
		)
		DO NOTHING
	`

	tag, err := tx.Exec(
		ctx,
		query,
		msg.Topic,
		msg.Partition,
		msg.Offset,
		msg.Key,
		msg.Value,
		result.Database,
		result.Table,
		result.Operation,
		entityID,
	)

	if err != nil {
		return false, fmt.Errorf("inserir evento CDC: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("confirmar transação: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
