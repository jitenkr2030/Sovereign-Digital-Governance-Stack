-- Migration Down: Drop all created objects

-- Drop triggers first
DROP TRIGGER IF EXISTS update_entities_updated_at ON compliance_entities;
DROP TRIGGER IF EXISTS update_regulations_updated_at ON compliance_regulations;
DROP TRIGGER IF EXISTS update_licenses_updated_at ON compliance_licenses;
DROP TRIGGER IF EXISTS update_applications_updated_at ON compliance_applications;
DROP TRIGGER IF EXISTS update_obligations_updated_at ON compliance_obligations;

-- Drop tables in order (respecting foreign keys)
DROP TABLE IF EXISTS compliance_audit_records CASCADE;
DROP TABLE IF EXISTS compliance_obligations CASCADE;
DROP TABLE IF EXISTS compliance_scores CASCADE;
DROP TABLE IF EXISTS compliance_applications CASCADE;
DROP TABLE IF EXISTS compliance_licenses CASCADE;
DROP TABLE IF EXISTS compliance_regulations CASCADE;
DROP TABLE IF EXISTS compliance_entities CASCADE;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop UUID extension (optional)
-- DROP EXTENSION IF EXISTS "uuid-ossp";
