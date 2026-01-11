-- Migration Down: Drop all created objects

-- Drop triggers
DROP TRIGGER IF EXISTS update_mining_operations_updated_at ON mining_operations;
DROP TRIGGER IF EXISTS update_mining_quotas_updated_at ON mining_quotas;

-- Drop tables in order (respecting foreign keys)
DROP TABLE IF EXISTS quota_violations CASCADE;
DROP TABLE IF EXISTS shutdown_commands CASCADE;
DROP TABLE IF EXISTS hashrate_records CASCADE;
DROP TABLE IF EXISTS mining_quotas CASCADE;
DROP TABLE IF EXISTS mining_operations CASCADE;

-- Drop enum types (if they exist and are not used elsewhere)
DROP TYPE IF EXISTS operation_status CASCADE;
DROP TYPE IF EXISTS command_status CASCADE;

-- Drop the UUID extension (optional, as other tables might use it)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
