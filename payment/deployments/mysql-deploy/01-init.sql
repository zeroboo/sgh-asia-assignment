-- Create the production database and load the schema.
CREATE DATABASE IF NOT EXISTS payment;
USE payment;

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id  VARCHAR(64)   NOT NULL,
    user_id         VARCHAR(64)   NOT NULL,
    amount          BIGINT NOT NULL DEFAULT 0,
    type            ENUM('debit','credit') NOT NULL,
    status          ENUM('created','processing','completed','failed','refunded') NOT NULL DEFAULT 'created',
    description     TEXT,
    created_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),

    PRIMARY KEY (transaction_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS user_balances (
    user_id    VARCHAR(64)   NOT NULL,
    balance    BIGINT        NOT NULL DEFAULT 0,
    updated_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS events (
    id              VARCHAR(128)  NOT NULL,
    transaction_id  VARCHAR(64)   NOT NULL,
    type            VARCHAR(32)   NOT NULL,
    payload         JSON,
    created_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS transaction_locks (
    transaction_id  VARCHAR(64)   NOT NULL,
    created_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    expires_at      DATETIME(3)   NOT NULL,
    deleted_at      DATETIME(3)   NULL,

    PRIMARY KEY (transaction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
