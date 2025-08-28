-- Financial Risk Monitor Database Setup Script
-- Run this script to create the database and user

-- Create database
CREATE DATABASE financial_risk_db;

-- Create user and grant permissions
CREATE USER riskmonitor WITH PASSWORD 'password123';
GRANT ALL PRIVILEGES ON DATABASE financial_risk_db TO riskmonitor;

-- Connect to the database
\c financial_risk_db;

-- Grant schema permissions
GRANT ALL ON SCHEMA public TO riskmonitor;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO riskmonitor;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO riskmonitor;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create enum types for better data consistency
DO $$ BEGIN
    CREATE TYPE alert_severity AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE alert_status AS ENUM ('ACTIVE', 'ACKNOWLEDGED', 'RESOLVED', 'DISMISSED');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE risk_status AS ENUM ('SAFE', 'WARNING', 'CRITICAL');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE liquidity_level AS ENUM ('HIGH', 'MEDIUM', 'LOW');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_portfolios_user_id ON portfolios(user_id);
CREATE INDEX IF NOT EXISTS idx_positions_portfolio_id ON positions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_id ON transactions(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_risk_metrics_portfolio_id ON risk_metrics(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_risk_metrics_calculated_at ON risk_metrics(calculated_at);
CREATE INDEX IF NOT EXISTS idx_risk_history_portfolio_id ON risk_history(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_risk_history_recorded_at ON risk_history(recorded_at);
CREATE INDEX IF NOT EXISTS idx_alerts_portfolio_id ON alerts(portfolio_id);
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status);
CREATE INDEX IF NOT EXISTS idx_alerts_created_at ON alerts(created_at);

-- Create composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_alerts_portfolio_status ON alerts(portfolio_id, status);
CREATE INDEX IF NOT EXISTS idx_risk_metrics_portfolio_type ON risk_metrics(portfolio_id, metric_type);
CREATE INDEX IF NOT EXISTS idx_transactions_portfolio_type ON transactions(portfolio_id, transaction_type);

COMMIT;

-- Display completion message
SELECT 'Database setup completed successfully!' as message;