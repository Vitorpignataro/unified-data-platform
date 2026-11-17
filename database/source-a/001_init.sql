
-- Source A: Cadastro

CREATE TABLE IF NOT EXISTS empresas (
    id_empresa BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    razao_social VARCHAR(180) NOT NULL,
    cnpj CHAR(14) NOT NULL,

    created_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6)
        ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id_empresa),
    UNIQUE KEY uk_empresas_cnpj (cnpj)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS unidades (
    id_unidade BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    id_empresa BIGINT UNSIGNED NOT NULL,

    nome VARCHAR(150) NOT NULL,
    cidade VARCHAR(120) NOT NULL,
    uf CHAR(2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ativa',

    created_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6),
    updated_at DATETIME(6) NOT NULL
        DEFAULT CURRENT_TIMESTAMP(6)
        ON UPDATE CURRENT_TIMESTAMP(6),

    PRIMARY KEY (id_unidade),

    KEY idx_unidades_empresa (id_empresa),
    KEY idx_unidades_status (status),

    CONSTRAINT fk_unidades_empresa
        FOREIGN KEY (id_empresa)
        REFERENCES empresas (id_empresa)
        ON DELETE RESTRICT,

    CONSTRAINT chk_unidades_status
        CHECK (status IN ('ativa', 'inativa'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
