import { MigrationInterface, QueryRunner } from "typeorm";

export class CreateLicensingSchema1712345600001 implements MigrationInterface {
    name = 'CreateLicensingSchema1712345600001'

    public async up(queryRunner: QueryRunner): Promise<void> {
        // Create license types table
        await queryRunner.query(`
            CREATE TABLE IF NOT EXISTS license_types (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                code VARCHAR(50) UNIQUE NOT NULL,
                name VARCHAR(100) NOT NULL,
                description TEXT,
                validity_months INTEGER NOT NULL DEFAULT 12,
                requirements JSONB DEFAULT '[]'::jsonb,
                metadata JSONB DEFAULT '{}'::jsonb,
                is_active BOOLEAN DEFAULT true,
                version INTEGER DEFAULT 1,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
            )
        `);

        // Create applications table
        await queryRunner.query(`
            CREATE TABLE IF NOT EXISTS applications (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                entity_id UUID NOT NULL,
                entity_name VARCHAR(255) NOT NULL,
                entity_type VARCHAR(100) NOT NULL,
                license_type_id UUID NOT NULL,
                status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
                priority VARCHAR(20) NOT NULL DEFAULT 'NORMAL',
                submission_data JSONB DEFAULT '{}'::jsonb,
                required_documents JSONB DEFAULT '[]'::jsonb,
                review_notes JSONB DEFAULT '[]'::jsonb,
                current_handler_id UUID,
                current_handler_name VARCHAR(255),
                submitted_at TIMESTAMP WITH TIME ZONE,
                review_started_at TIMESTAMP WITH TIME ZONE,
                decision_at TIMESTAMP WITH TIME ZONE,
                decision VARCHAR(50),
                decision_reason TEXT,
                due_date TIMESTAMP WITH TIME ZONE,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                CONSTRAINT fk_license_type FOREIGN KEY (license_type_id)
                    REFERENCES license_types(id) ON DELETE RESTRICT
            )
        `);

        // Create licenses table
        await queryRunner.query(`
            CREATE TABLE IF NOT EXISTS licenses (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                application_id UUID NOT NULL,
                entity_id UUID NOT NULL,
                entity_name VARCHAR(255) NOT NULL,
                entity_type VARCHAR(100) NOT NULL,
                license_number VARCHAR(50) UNIQUE NOT NULL,
                license_type_id UUID NOT NULL,
                license_type_name VARCHAR(100) NOT NULL,
                status VARCHAR(50) NOT NULL DEFAULT 'ACTIVE',
                issue_date DATE NOT NULL,
                expiry_date DATE NOT NULL,
                renewal_window_start DATE,
                renewed_at TIMESTAMP WITH TIME ZONE,
                suspended_at TIMESTAMP WITH TIME ZONE,
                revoked_at TIMESTAMP WITH TIME ZONE,
                suspended_reason TEXT,
                revoked_reason TEXT,
                suspended_by VARCHAR(255),
                revoked_by VARCHAR(255),
                scope JSONB DEFAULT '[]'::jsonb,
                conditions JSONB DEFAULT '[]'::jsonb,
                metadata JSONB DEFAULT '{}'::jsonb,
                certificate_url VARCHAR(500),
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                CONSTRAINT fk_application FOREIGN KEY (application_id)
                    REFERENCES applications(id) ON DELETE RESTRICT
            )
        `);

        // Create documents table
        await queryRunner.query(`
            CREATE TABLE IF NOT EXISTS documents (
                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                application_id UUID NOT NULL,
                file_name VARCHAR(500) NOT NULL,
                original_name VARCHAR(255) NOT NULL,
                mime_type VARCHAR(100) NOT NULL,
                file_size BIGINT NOT NULL,
                s3_key VARCHAR(500) NOT NULL,
                s3_bucket VARCHAR(100),
                document_type VARCHAR(50) NOT NULL,
                status VARCHAR(50) NOT NULL DEFAULT 'UPLOADED',
                checksum VARCHAR(64) NOT NULL,
                checksum_algorithm VARCHAR(20) DEFAULT 'sha256',
                description TEXT,
                uploaded_by UUID,
                uploaded_by_name VARCHAR(255),
                verified_at TIMESTAMP WITH TIME ZONE,
                verified_by UUID,
                verified_by_name VARCHAR(255),
                verification_notes TEXT,
                rejected_at TIMESTAMP WITH TIME ZONE,
                rejection_reason TEXT,
                metadata JSONB DEFAULT '{}'::jsonb,
                is_confidential BOOLEAN DEFAULT false,
                quarantine_until TIMESTAMP WITH TIME ZONE,
                created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                CONSTRAINT fk_application_doc FOREIGN KEY (application_id)
                    REFERENCES applications(id) ON DELETE CASCADE
            )
        `);

        // Create indexes for better query performance
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_applications_entity_id ON applications(entity_id)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_applications_handler ON applications(current_handler_id)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_licenses_status ON licenses(status)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_licenses_entity_id ON licenses(entity_id)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_licenses_expiry ON licenses(expiry_date)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_documents_application ON documents(application_id)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_documents_status ON documents(status)`);
        await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(document_type)`);

        // Seed default license types
        await queryRunner.query(`
            INSERT INTO license_types (code, name, description, validity_months, requirements)
            VALUES
                ('VASP', 'Virtual Asset Service Provider', 'License for entities providing virtual asset services', 12,
                 '["proof_of_incorporation", "aml_kyc_policy", "business_plan", "financial_statements", "security_audit_report"]'),
                ('CASP', 'Crypto Asset Service Provider', 'License for crypto asset service providers', 12,
                 '["proof_of_incorporation", "aml_kyc_policy", "operational_reserve_proof", "security_audit_report"]'),
                ('EXCHANGE', 'Cryptocurrency Exchange', 'License for operating cryptocurrency exchanges', 12,
                 '["proof_of_incorporation", "aml_kyc_policy", "business_plan", "financial_statements", "security_audit_report", "cold_storage_policy", "insurance_proof"]'),
                ('WALLET', 'Custodial Wallet Provider', 'License for custodial wallet service providers', 24,
                 '["proof_of_incorporation", "aml_kyc_policy", "security_audit_report", "custody_policy"]'),
                ('MINING', 'Mining Pool Operator', 'License for mining pool operators', 24,
                 '["proof_of_incorporation", "energy_source_documentation", "pool_software_security_audit", "geolocation_data"]')
            ON CONFLICT (code) DO NOTHING
        `);
    }

    public async down(queryRunner: QueryRunner): Promise<void> {
        await queryRunner.query(`DROP TABLE IF EXISTS documents CASCADE`);
        await queryRunner.query(`DROP TABLE IF EXISTS licenses CASCADE`);
        await queryRunner.query(`DROP TABLE IF EXISTS applications CASCADE`);
        await queryRunner.query(`DROP TABLE IF EXISTS license_types CASCADE`);
    }
}
