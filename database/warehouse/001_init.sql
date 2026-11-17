
BEGIN;

CREATE SCHEMA IF NOT EXISTS raw;
CREATE SCHEMA IF NOT EXISTS core;
CREATE SCHEMA IF NOT EXISTS analytics;

-- Eventos recebidos do CDC

CREATE TABLE IF NOT EXISTS raw.cdc_events (
    event_id TEXT PRIMARY KEY,
    source_system VARCHAR(32) NOT NULL,
    source_table VARCHAR(64) NOT NULL,
    source_pk VARCHAR(128) NOT NULL,
    operation CHAR(1) NOT NULL,
    payload JSONB NOT NULL,
    source_timestamp TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_cdc_operation
        CHECK (operation IN ('r', 'c', 'u', 'd'))
);

CREATE INDEX IF NOT EXISTS idx_cdc_source_key
    ON raw.cdc_events (
        source_system,
        source_table,
        source_pk
    );

-- Empresas unificadas

CREATE TABLE IF NOT EXISTS core.empresas (
    id_empresa BIGINT PRIMARY KEY,
    razao_social VARCHAR(180) NOT NULL,
    cnpj CHAR(14) NOT NULL UNIQUE,

    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Unidades unificadas

CREATE TABLE IF NOT EXISTS core.unidades (
    id_unidade BIGINT PRIMARY KEY,
    id_empresa BIGINT NOT NULL
        REFERENCES core.empresas(id_empresa),

    nome VARCHAR(150) NOT NULL,
    cidade VARCHAR(120) NOT NULL,
    uf CHAR(2) NOT NULL,
    status VARCHAR(20) NOT NULL,

    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_core_unidade_status
        CHECK (status IN ('ativa', 'inativa'))
);

CREATE INDEX IF NOT EXISTS idx_core_unidades_empresa
    ON core.unidades (id_empresa);

-- Faturas unificadas

CREATE TABLE IF NOT EXISTS core.faturas (
    id_fatura BIGINT PRIMARY KEY,
    id_unidade BIGINT NOT NULL
        REFERENCES core.unidades(id_unidade),

    mes_referencia DATE NOT NULL,
    valor_total NUMERIC(14,2) NOT NULL,
    status VARCHAR(20) NOT NULL,

    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_core_fatura_valor
        CHECK (valor_total >= 0),

    CONSTRAINT chk_core_fatura_status
        CHECK (
            status IN ('pendente', 'paga', 'cancelada')
        )
);

CREATE INDEX IF NOT EXISTS idx_core_faturas_unidade_mes
    ON core.faturas (id_unidade, mes_referencia);

CREATE INDEX IF NOT EXISTS idx_core_faturas_status
    ON core.faturas (status);

-- Pagamentos unificados

CREATE TABLE IF NOT EXISTS core.pagamentos (
    id_pagamento BIGINT PRIMARY KEY,
    id_fatura BIGINT NOT NULL
        REFERENCES core.faturas(id_fatura),

    valor_pago NUMERIC(14,2) NOT NULL
        CHECK (valor_pago > 0),
    data_pagamento DATE NOT NULL,

    source_updated_at TIMESTAMPTZ,
    synced_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_core_pagamentos_fatura
    ON core.pagamentos (id_fatura);

-- View de consulta integrada

CREATE OR REPLACE VIEW analytics.vw_faturas_por_unidade AS
SELECT
    e.id_empresa,
    e.razao_social,
    u.id_unidade,
    u.nome AS unidade,
    u.cidade,
    u.uf,
    f.id_fatura,
    f.mes_referencia,
    f.valor_total,
    f.status AS status_fatura
FROM core.empresas e
INNER JOIN core.unidades u
    ON u.id_empresa = e.id_empresa
INNER JOIN core.faturas f
    ON f.id_unidade = u.id_unidade;

COMMIT;
