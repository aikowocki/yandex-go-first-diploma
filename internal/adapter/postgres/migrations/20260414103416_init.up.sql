CREATE TYPE order_status AS ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
CREATE TYPE transaction_type AS ENUM ('ACCRUAL', 'WITHDRAWAL');


CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    login varchar(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    number VARCHAR(255) NOT NULL UNIQUE,
    status order_status NOT NUll DEFAULT 'NEW',
    accrual BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    order_number VARCHAR(255) NOT NULL,
    type transaction_type NOT NULL,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status On orders(status);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type on transactions(user_id, type);
CREATE UNIQUE INDEX idx_transactions_unique_operation ON transactions(user_id, order_number, type)