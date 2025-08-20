CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL,
    symbol VARCHAR(10),
    quantity DECIMAL(20,8),
    price DECIMAL(20,8),
    amount DECIMAL(20,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'PENDING',
    executed_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    kyc_verified BOOLEAN DEFAULT false,
    aml_checked BOOLEAN DEFAULT false,
    risk_score INTEGER,
    compliance_notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_id ON transactions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
