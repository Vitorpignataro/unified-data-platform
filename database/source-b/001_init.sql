
-- Source B: Financeiro

CREATE TABLE IF NOT EXISTS faturas (
    id_fatura BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    id_unidade BIGINT UNSIGNED NOT NULL,

    mes_referencia DATE NOT NULL,
    valor_total DECIMAL(14,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pendente',

    created_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6)
        ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id_fatura),

    KEY idx_faturas_unidade_mes (
        id_unidade,
        mes_referencia
    ),

    KEY idx_faturas_status (status),

    CONSTRAINT chk_faturas_valor
        CHECK (valor_total >= 0),

    CONSTRAINT chk_faturas_status
        CHECK (
            status IN (
                'pendente',
                'paga',
                'cancelada'
            )
        )
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS pagamentos (
    id_pagamento BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    id_fatura BIGINT UNSIGNED NOT NULL,

    valor_pago DECIMAL(14,2) NOT NULL,
    data_pagamento DATE NOT NULL,

    created_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6)
        ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id_pagamento),

    KEY idx_pagamentos_fatura (id_fatura),

    CONSTRAINT fk_pagamentos_fatura
        FOREIGN KEY (id_fatura)
        REFERENCES faturas (id_fatura)
        ON DELETE RESTRICT,

    CONSTRAINT chk_pagamentos_valor
        CHECK (valor_pago > 0)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
