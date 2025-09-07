-- Rollback for Alpaca fields on users
ALTER TABLE users
DROP COLUMN IF EXISTS alpaca_account_id,
DROP COLUMN IF EXISTS alpaca_access_token,
DROP COLUMN IF EXISTS alpaca_refresh_token,
DROP COLUMN IF EXISTS alpaca_expires_at,
DROP COLUMN IF EXISTS alpaca_is_linked;



