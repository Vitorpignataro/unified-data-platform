BEGIN;

CREATE SCHEMA IF NOT EXISTS raw;

CREATE TABLE IF NOT EXISTS raw.kafka_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- Identificação da mensagem Kafka
    topic TEXT NOT NULL,
    partition_id INTEGER NOT NULL,
    offset_id BIGINT NOT NULL,

    -- Conteúdo original
    message_key BYTEA,
    message_value BYTEA,

    -- Metadados interpretados
    source_database VARCHAR(100) NOT NULL,
    source_table VARCHAR(100) NOT NULL,
    operation VARCHAR(16) NOT NULL,
    entity_id BIGINT,

    -- Momento da persistência
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Impede mensagens duplicadas
    CONSTRAINT uq_kafka_message_position
        UNIQUE (
            topic,
            partition_id,
            offset_id
        ),

    CONSTRAINT chk_kafka_partition
        CHECK (partition_id >= 0),

    CONSTRAINT chk_kafka_offset
        CHECK (offset_id >= 0),

    CONSTRAINT chk_cdc_operation
        CHECK (
            operation IN (
                'r',
                'c',
                'u',
                'd',
                'tombstone'
            )
        ),

    CONSTRAINT chk_message_value
        CHECK (
            (
                operation = 'tombstone'
                AND message_value IS NULL
            )
            OR (
                operation <> 'tombstone'
                AND message_value IS NOT NULL
            )
        ),

    CONSTRAINT chk_entity_id
        CHECK (entity_id IS NULL OR entity_id > 0)
);

CREATE INDEX IF NOT EXISTS idx_kafka_messages_source
    ON raw.kafka_messages (
        source_database,
        source_table,
        entity_id
    );

CREATE INDEX IF NOT EXISTS idx_kafka_messages_received
    ON raw.kafka_messages (received_at);

COMMIT;
