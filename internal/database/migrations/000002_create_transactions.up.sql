CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    hash VARCHAR(66) UNIQUE NOT NULL,
    sender VARCHAR(42) NOT NULL,
    receiver VARCHAR(42) NOT NULL,
    amount DECIMAL(30,18) NOT NULL,
    timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    status tx_status DEFAULT 'pending' NOT NULL
    );


CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender);
CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver);