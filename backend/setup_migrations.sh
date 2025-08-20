#!/bin/bash

# Create migrations directory
mkdir -p migrations

# Create 001_create_users.up.sql
cat > migrations/001_create_users.up.sql << 'EOF'
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'analyst',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
EOF

# Create 001_create_users.down.sql
cat > migrations/001_create_users.down.sql << 'EOF'
DROP TABLE IF EXISTS users;
EOF

# Create 002_create_portfolios.up.sql
cat > migrations/002_create_portfolios.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_value DECIMAL(20,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_portfolios_user_id ON portfolios(user_id);
EOF

# Create 002_create_portfolios.down.sql
cat > migrations/002_create_portfolios.down.sql << 'EOF'
DROP TABLE IF EXISTS portfolios;
EOF

# Create 003_create_positions.up.sql
cat > migrations/003_create_positions.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    symbol VARCHAR(10) NOT NULL,
    quantity DECIMAL(20,8) NOT NULL,
    average_price DECIMAL(20,8) NOT NULL,
    current_price DECIMAL(20,8) NOT NULL,
    market_value DECIMAL(20,2) NOT NULL,
    pn_l DECIMAL(20,2),
    pn_l_percent DECIMAL(10,4),
    weight DECIMAL(10,4),
    asset_type VARCHAR(50) NOT NULL,
    liquidity VARCHAR(20) DEFAULT 'HIGH',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_positions_portfolio_id ON positions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_positions_symbol ON positions(symbol);
EOF

# Create 003_create_positions.down.sql
cat > migrations/003_create_positions.down.sql << 'EOF'
DROP TABLE IF EXISTS positions;
EOF

# Create 004_create_transactions.up.sql
cat > migrations/004_create_transactions.up.sql << 'EOF'
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
EOF

# Create 004_create_transactions.down.sql
cat > migrations/004_create_transactions.down.sql << 'EOF'
DROP TABLE IF EXISTS transactions;
EOF

# Create 005_create_risk_metrics.up.sql
cat > migrations/005_create_risk_metrics.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS risk_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,
    value DECIMAL(20,8) NOT NULL,
    threshold DECIMAL(20,8),
    status VARCHAR(20) NOT NULL,
    calculated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    time_horizon INTEGER,
    confidence_level DECIMAL(5,4),
    details JSONB
);

CREATE TABLE IF NOT EXISTS risk_histories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    metric_type VARCHAR(50) NOT NULL,
    value DECIMAL(20,8) NOT NULL,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_risk_metrics_portfolio_id ON risk_metrics(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_risk_metrics_metric_type ON risk_metrics(metric_type);
CREATE INDEX IF NOT EXISTS idx_risk_histories_portfolio_id ON risk_histories(portfolio_id);
EOF

# Create 005_create_risk_metrics.down.sql
cat > migrations/005_create_risk_metrics.down.sql << 'EOF'
DROP TABLE IF EXISTS risk_histories;
DROP TABLE IF EXISTS risk_metrics;
EOF

# Create 006_create_alerts.up.sql
cat > migrations/006_create_alerts.up.sql << 'EOF'
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(20) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    source VARCHAR(100),
    status VARCHAR(20) DEFAULT 'ACTIVE',
    triggered_by JSONB,
    resolution TEXT,
    acknowledged_by UUID REFERENCES users(id),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alerts_portfolio_id ON alerts(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);
EOF

# Create 006_create_alerts.down.sql
cat > migrations/006_create_alerts.down.sql << 'EOF'
DROP TABLE IF EXISTS alerts;
EOF

echo "✅ Migration files created successfully!"
echo ""
echo "Files created in ./migrations/:"
ls -la migrations/