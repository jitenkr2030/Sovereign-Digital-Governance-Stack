-- Migration V1: Create Initial Schema for Wallet Governance
-- Direction: UP

-- Create ENUM types
CREATE TYPE wallet_type AS ENUM (
	'CUSTODIAL', 'MULTI_SIG', 'EXCHANGE_HOT', 'EXCHANGE_COLD',
	'TREASURY', 'ESCROW', 'USER_CONTROLLED'
);

CREATE TYPE wallet_status AS ENUM (
	'ACTIVE', 'SUSPENDED', 'FROZEN', 'REVOKED', 'PENDING', 'UNDER_REVIEW'
);

CREATE TYPE blockchain_type AS ENUM (
	'BITCOIN', 'ETHEREUM', 'ERC20', 'TRC20', 'BEP20', 'POLYGON',
	'SOLANA', 'LITECOIN', 'DASH', 'MONERO'
);

CREATE TYPE signer_type AS ENUM ('INSTITUTION', 'INDIVIDUAL', 'AUTOMATED', 'EMERGENCY');
CREATE TYPE signer_status AS ENUM ('ACTIVE', 'INACTIVE', 'REVOKED', 'PENDING');
CREATE TYPE transaction_status AS ENUM (
	'PENDING', 'APPROVED', 'REJECTED', 'EXECUTED', 'FAILED', 'EXPIRED', 'CANCELLED'
);
CREATE TYPE signature_status AS ENUM (
	'PENDING', 'IN_PROGRESS', 'COMPLETED', 'FAILED', 'CANCELLED'
);
CREATE TYPE freeze_status AS ENUM ('ACTIVE', 'PARTIAL', 'RELEASED', 'EXPIRED');
CREATE TYPE freeze_reason AS ENUM (
	'LEGAL_ORDER', 'REGULATORY', 'SUSPICIOUS_ACTIVITY', 'EXCHANGE_HACK',
	'USER_REQUEST', 'MAINTENANCE'
);
CREATE TYPE blacklist_status AS ENUM ('ACTIVE', 'EXPIRED', 'REMOVED', 'UNDER_REVIEW');

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id VARCHAR(50) NOT NULL UNIQUE,
	type wallet_type NOT NULL,
	status wallet_status NOT NULL DEFAULT 'ACTIVE',
	blockchain blockchain_type NOT NULL,
	address VARCHAR(255) NOT NULL,
	address_checksum VARCHAR(64),
	label VARCHAR(255),
	description TEXT,
	exchange_id UUID,
	exchange_name VARCHAR(255),
	owner_entity_id UUID NOT NULL,
	owner_entity_name VARCHAR(255) NOT NULL,
	signers_required INT NOT NULL DEFAULT 1,
	signers_total INT NOT NULL DEFAULT 0,
	threshold INT NOT NULL DEFAULT 1,
	total_balance DECIMAL(30, 8) NOT NULL DEFAULT 0,
	balance_currency VARCHAR(10) NOT NULL DEFAULT 'USD',
	last_activity_at TIMESTAMP WITH TIME ZONE,
	compliance_score DECIMAL(5, 2) NOT NULL DEFAULT 100.00,
	is_whitelisted BOOLEAN NOT NULL DEFAULT FALSE,
	is_blacklisted BOOLEAN NOT NULL DEFAULT FALSE,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_wallets_address ON wallets(address, blockchain);
CREATE INDEX IF NOT EXISTS idx_wallets_wallet_id ON wallets(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallets_exchange_id ON wallets(exchange_id);
CREATE INDEX IF NOT EXISTS idx_wallets_owner_entity ON wallets(owner_entity_id);
CREATE INDEX IF NOT EXISTS idx_wallets_status ON wallets(status);
CREATE INDEX IF NOT EXISTS idx_wallets_type ON wallets(type);
CREATE INDEX IF NOT EXISTS idx_wallets_blockchain ON wallets(blockchain);

-- Wallet signers table
CREATE TABLE IF NOT EXISTS wallet_signers (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
	signer_id VARCHAR(100) NOT NULL,
	signer_type signer_type NOT NULL,
	signer_name VARCHAR(255) NOT NULL,
	public_key TEXT NOT NULL,
	public_key_hash VARCHAR(64) NOT NULL,
	status signer_status NOT NULL DEFAULT 'ACTIVE',
	order_num INT NOT NULL DEFAULT 0,
	is_emergency BOOLEAN NOT NULL DEFAULT FALSE,
	weight INT NOT NULL DEFAULT 1,
	last_signed_at TIMESTAMP WITH TIME ZONE,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT unique_signer UNIQUE (wallet_id, signer_id)
);

CREATE INDEX IF NOT EXISTS idx_wallet_signers_wallet ON wallet_signers(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_signers_signer ON wallet_signers(signer_id);
CREATE INDEX IF NOT EXISTS idx_wallet_signers_status ON wallet_signers(status);

-- Transaction proposals table
CREATE TABLE IF NOT EXISTS transaction_proposals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
	transaction_id VARCHAR(50) NOT NULL UNIQUE,
	tx_type VARCHAR(50) NOT NULL,
	blockchain blockchain_type NOT NULL,
	to_address VARCHAR(255) NOT NULL,
	amount DECIMAL(30, 8) NOT NULL,
	asset_symbol VARCHAR(20) NOT NULL,
	contract_address VARCHAR(255),
	gas_limit INT NOT NULL DEFAULT 21000,
	gas_price DECIMAL(30, 8),
	nonce INT,
	raw_transaction TEXT,
	signed_transactions TEXT[],
	status transaction_status NOT NULL DEFAULT 'PENDING',
	proposer_id UUID NOT NULL,
	proposer_name VARCHAR(255) NOT NULL,
	approvers_required INT NOT NULL DEFAULT 1,
	approvals_count INT NOT NULL DEFAULT 0,
	rejections_count INT NOT NULL DEFAULT 0,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	executed_at TIMESTAMP WITH TIME ZONE,
	failure_reason TEXT,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_proposals_wallet ON transaction_proposals(wallet_id);
CREATE INDEX IF NOT EXISTS idx_tx_proposals_status ON transaction_proposals(status);
CREATE INDEX IF NOT EXISTS idx_tx_proposals_expires ON transaction_proposals(expires_at);
CREATE INDEX IF NOT EXISTS idx_tx_proposals_proposer ON transaction_proposals(proposer_id);

-- Transaction approvals table
CREATE TABLE IF NOT EXISTS transaction_approvals (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	transaction_id UUID NOT NULL REFERENCES transaction_proposals(id) ON DELETE CASCADE,
	signer_id UUID NOT NULL,
	signer_name VARCHAR(255) NOT NULL,
	decision VARCHAR(20) NOT NULL,
	reason TEXT,
	signature TEXT,
	signature_expiry TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_approvals_tx ON transaction_approvals(transaction_id);

-- Signature requests table
CREATE TABLE IF NOT EXISTS signature_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	request_id VARCHAR(50) NOT NULL UNIQUE,
	wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
	signer_id UUID NOT NULL,
	message_hash VARCHAR(64) NOT NULL,
	message TEXT NOT NULL,
	signature_type VARCHAR(50) NOT NULL,
	status signature_status NOT NULL DEFAULT 'PENDING',
	signature TEXT,
	public_key TEXT,
	expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
	completed_at TIMESTAMP WITH TIME ZONE,
	failure_reason TEXT,
	retry_count INT NOT NULL DEFAULT 0,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sig_requests_wallet ON signature_requests(wallet_id);
CREATE INDEX IF NOT EXISTS idx_sig_requests_status ON signature_requests(status);
CREATE INDEX IF NOT EXISTS idx_sig_requests_expires ON signature_requests(expires_at);

-- Blacklist table
CREATE TABLE IF NOT EXISTS blacklist (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	address VARCHAR(255) NOT NULL,
	address_hash VARCHAR(64) NOT NULL,
	blockchain blockchain_type NOT NULL,
	reason VARCHAR(100) NOT NULL,
	reason_details TEXT,
	source VARCHAR(100) NOT NULL,
	risk_level VARCHAR(20) NOT NULL,
	status blacklist_status NOT NULL DEFAULT 'ACTIVE',
	expires_at TIMESTAMP WITH TIME ZONE,
	added_by UUID NOT NULL,
	added_by_name VARCHAR(255) NOT NULL,
	approved_by UUID,
	removed_by UUID,
	removal_reason TEXT,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT unique_blacklist_address UNIQUE (address, blockchain)
);

CREATE INDEX IF NOT EXISTS idx_blacklist_address ON blacklist(address, blockchain);
CREATE INDEX IF NOT EXISTS idx_blacklist_status ON blacklist(status);
CREATE INDEX IF NOT EXISTS idx_blacklist_source ON blacklist(source);
CREATE INDEX IF NOT EXISTS idx_blacklist_risk ON blacklist(risk_level);

-- Whitelist table
CREATE TABLE IF NOT EXISTS whitelist (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	address VARCHAR(255) NOT NULL,
	address_hash VARCHAR(64) NOT NULL,
	blockchain blockchain_type NOT NULL,
	label VARCHAR(255),
	description TEXT,
	entity_type VARCHAR(50) NOT NULL,
	entity_id UUID,
	entity_name VARCHAR(255),
	status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
	expires_at TIMESTAMP WITH TIME ZONE,
	added_by UUID NOT NULL,
	added_by_name VARCHAR(255) NOT NULL,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	CONSTRAINT unique_whitelist_address UNIQUE (address, blockchain)
);

CREATE INDEX IF NOT EXISTS idx_whitelist_address ON whitelist(address, blockchain);
CREATE INDEX IF NOT EXISTS idx_whitelist_status ON whitelist(status);
CREATE INDEX IF NOT EXISTS idx_whitelist_entity ON whitelist(entity_id);

-- Wallet freezes table
CREATE TABLE IF NOT EXISTS wallet_freezes (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
	wallet_address VARCHAR(255) NOT NULL,
	blockchain blockchain_type NOT NULL,
	reason freeze_reason NOT NULL,
	reason_details TEXT,
	status freeze_status NOT NULL DEFAULT 'ACTIVE',
	freeze_level VARCHAR(20) NOT NULL DEFAULT 'FULL',
	legal_order_id VARCHAR(100),
	issued_by UUID NOT NULL,
	issued_by_name VARCHAR(255) NOT NULL,
	approved_by UUID,
	expires_at TIMESTAMP WITH TIME ZONE,
	released_at TIMESTAMP WITH TIME ZONE,
	release_reason TEXT,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_freezes_wallet ON wallet_freezes(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_freezes_status ON wallet_freezes(status);
CREATE INDEX IF NOT EXISTS idx_wallet_freezes_issued ON wallet_freezes(issued_by);
CREATE INDEX IF NOT EXISTS idx_wallet_freezes_expires ON wallet_freezes(expires_at);

-- Asset recovery requests table
CREATE TABLE IF NOT EXISTS asset_recovery_requests (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	request_id VARCHAR(50) NOT NULL UNIQUE,
	wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
	source_blockchain blockchain_type NOT NULL,
	source_address VARCHAR(255) NOT NULL,
	target_address VARCHAR(255) NOT NULL,
	asset_symbol VARCHAR(20) NOT NULL,
	amount DECIMAL(30, 8) NOT NULL,
	reason TEXT NOT NULL,
	legal_case_id VARCHAR(100),
	status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
	requester_id UUID NOT NULL,
	requester_name VARCHAR(255) NOT NULL,
	approver_id UUID,
	tx_hash VARCHAR(255),
	failure_reason TEXT,
	metadata JSONB,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
	executed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_recovery_wallet ON asset_recovery_requests(wallet_id);
CREATE INDEX IF NOT EXISTS idx_recovery_status ON asset_recovery_requests(status);
CREATE INDEX IF NOT EXISTS idx_recovery_requester ON asset_recovery_requests(requester_id);

-- Wallet audit logs table
CREATE TABLE IF NOT EXISTS wallet_audit_logs (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	entity_type VARCHAR(100) NOT NULL,
	entity_id UUID NOT NULL,
	action VARCHAR(50) NOT NULL,
	actor_id UUID NOT NULL,
	actor_name VARCHAR(255) NOT NULL,
	actor_type VARCHAR(50) NOT NULL,
	old_value JSONB,
	new_value JSONB,
	ip_address INET,
	user_agent TEXT,
	request_id VARCHAR(100),
	success BOOLEAN NOT NULL DEFAULT TRUE,
	error_message TEXT,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity ON wallet_audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON wallet_audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON wallet_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON wallet_audit_logs(created_at DESC);

-- Direction: DOWN
-- DROP TABLE IF EXISTS wallet_audit_logs CASCADE;
-- DROP TABLE IF EXISTS asset_recovery_requests CASCADE;
-- DROP TABLE IF EXISTS wallet_freezes CASCADE;
-- DROP TABLE IF EXISTS whitelist CASCADE;
-- DROP TABLE IF EXISTS blacklist CASCADE;
-- DROP TABLE IF EXISTS signature_requests CASCADE;
-- DROP TABLE IF EXISTS transaction_approvals CASCADE;
-- DROP TABLE IF EXISTS transaction_proposals CASCADE;
-- DROP TABLE IF EXISTS wallet_signers CASCADE;
-- DROP TABLE IF EXISTS wallets CASCADE;
-- DROP TYPE IF EXISTS blacklist_status CASCADE;
-- DROP TYPE IF EXISTS freeze_reason CASCADE;
-- DROP TYPE IF EXISTS freeze_status CASCADE;
-- DROP TYPE IF EXISTS signature_status CASCADE;
-- DROP TYPE IF EXISTS transaction_status CASCADE;
-- DROP TYPE IF EXISTS signer_status CASCADE;
-- DROP TYPE IF EXISTS signer_type CASCADE;
-- DROP TYPE IF EXISTS blockchain_type CASCADE;
-- DROP TYPE IF EXISTS wallet_status CASCADE;
-- DROP TYPE IF EXISTS wallet_type CASCADE;
