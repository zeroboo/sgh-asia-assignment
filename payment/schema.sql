CREATE TABLE transactions (
    transaction_id  VARCHAR(64)   NOT NULL,
    user_id         VARCHAR(64)   NOT NULL,
    amount          DECIMAL(18,2) NOT NULL,
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

CREATE TABLE user_balances (
    user_id    VARCHAR(64)   NOT NULL,
    balance    BIGINT        NOT NULL DEFAULT 0,
    updated_at DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE events (
    id              VARCHAR(128)  NOT NULL,
    transaction_id  VARCHAR(64)   NOT NULL,
    type            VARCHAR(32)   NOT NULL,  -- e.g. 'payment.created', 'payment.completed'
    payload         JSON,
    created_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),

    PRIMARY KEY (id),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_type (type),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE transaction_locks (
    transaction_id  VARCHAR(64)   NOT NULL,
    created_at      DATETIME(3)   NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    deleted_at      DATETIME(3)   NULL,

    PRIMARY KEY (transaction_id),
    CONSTRAINT fk_lock_transaction FOREIGN KEY (transaction_id) REFERENCES transactions(transaction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;