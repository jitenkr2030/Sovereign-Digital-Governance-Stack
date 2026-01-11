/**
 * Configuration Management
 * 
 * Loads configuration from config.yaml with environment variable override support.
 */

import * as fs from 'fs';
import * as path from 'path';
import * as yaml from 'yaml';

export interface AppConfig {
  name: string;
  environment: string;
  version: string;
}

export interface ServerConfig {
  host: string;
  port: number;
  read_timeout: number;
  write_timeout: number;
  idle_timeout: number;
  corsOrigins: string[];
}

export interface DatabaseConfig {
  host: string;
  port: number;
  username: string;
  password: string;
  name: string;
  ssl_mode: string;
  max_open_conns: number;
  max_idle_conns: number;
  conn_max_lifetime: number;
}

export interface RedisConfig {
  host: string;
  port: number;
  password: string;
  db: number;
  pool_size: number;
  key_prefix: string;
}

export interface QueueConfig {
  name: string;
  concurrency: number;
  retry_attempts: number;
  retry_delay: number;
}

export interface KafkaConfig {
  brokers: string[];
  client_id: string;
  group_id: string;
  topics: {
    report_generated: string;
    report_failed: string;
  };
}

export interface S3Config {
  endpoint: string;
  access_key: string;
  secret_key: string;
  bucket: string;
  region: string;
  signed_url_expiry: number;
}

export interface WORMConfig {
  enabled: boolean;
  retention_days: number;
  retention_mode: string;
  legal_hold_enabled: boolean;
}

export interface StorageConfig {
  provider: string;
  s3: S3Config;
  worm: WORMConfig;
}

export interface ReportTypeConfig {
  id: string;
  name: string;
  description: string;
  template: string;
  format: string;
  schedule_cron: string;
}

export interface PDFConfig {
  default_page_size: string;
  default_margin: number;
  header_height: number;
  footer_height: number;
  font_family: string;
  font_size: number;
}

export interface CSVConfig {
  delimiter: string;
  encoding: string;
  include_headers: boolean;
  date_format: string;
  datetime_format: string;
}

export interface RetentionConfig {
  max_reports_per_type: number;
  archive_after_days: number;
  delete_after_days: number;
}

export interface ReportingConfig {
  types: ReportTypeConfig[];
  pdf: PDFConfig;
  csv: CSVConfig;
  retention: RetentionConfig;
}

export interface PythonWorkerConfig {
  enabled: boolean;
  host: string;
  port: number;
  max_workers: number;
  timeout_seconds: number;
}

export interface JWTAuthConfig {
  secret: string;
  expiry: string;
}

export interface RateLimitConfig {
  enabled: boolean;
  requests_per_minute: number;
  burst_size: number;
}

export interface SecurityConfig {
  jwt: JWTAuthConfig;
  rate_limit: RateLimitConfig;
}

export interface LoggingConfig {
  level: string;
  format: string;
  include_caller: boolean;
}

export interface MetricsConfig {
  enabled: boolean;
  endpoint: string;
  port: number;
  prefix: string;
}

export interface Config {
  app: AppConfig;
  server: ServerConfig;
  database: DatabaseConfig;
  redis: RedisConfig;
  queue: QueueConfig;
  kafka: KafkaConfig;
  storage: StorageConfig;
  reporting: ReportingConfig;
  python_worker: PythonWorkerConfig;
  security: SecurityConfig;
  logging: LoggingConfig;
  metrics: MetricsConfig;
}

let cachedConfig: Config | null = null;

function loadConfig(): Config {
  if (cachedConfig) {
    return cachedConfig;
  }

  // Look for config.yaml in various locations
  const possiblePaths = [
    path.join(process.cwd(), 'config.yaml'),
    path.join(process.cwd(), '..', 'config.yaml'),
    path.join(process.cwd(), '..', '..', 'config.yaml'),
    path.join(__dirname, '..', 'config.yaml'),
    path.join(__dirname, '..', 'internal', 'config', 'config.yaml'),
  ];

  let configData: any = {};

  for (const configPath of possiblePaths) {
    if (fs.existsSync(configPath)) {
      const fileContent = fs.readFileSync(configPath, 'utf8');
      configData = yaml.parse(fileContent);
      break;
    }
  }

  // Apply environment variable overrides
  configData = applyEnvOverrides(configData);

  // Validate required fields
  validateConfig(configData);

  cachedConfig = configData as Config;
  return cachedConfig;
}

function applyEnvOverrides(config: any): any {
  // Server overrides
  if (process.env.PORT) {
    config.server = config.server || {};
    config.server.port = parseInt(process.env.PORT, 10);
  }
  if (process.env.NODE_ENV) {
    config.app = config.app || {};
    config.app.environment = process.env.NODE_ENV;
  }

  // Database overrides
  if (process.env.DB_HOST) {
    config.database = config.database || {};
    config.database.host = process.env.DB_HOST;
  }
  if (process.env.DB_PORT) {
    config.database = config.database || {};
    config.database.port = parseInt(process.env.DB_PORT, 10);
  }
  if (process.env.DB_USERNAME) {
    config.database = config.database || {};
    config.database.username = process.env.DB_USERNAME;
  }
  if (process.env.DB_PASSWORD) {
    config.database = config.database || {};
    config.database.password = process.env.DB_PASSWORD;
  }
  if (process.env.DB_NAME) {
    config.database = config.database || {};
    config.database.name = process.env.DB_NAME;
  }

  // Redis overrides
  if (process.env.REDIS_HOST) {
    config.redis = config.redis || {};
    config.redis.host = process.env.REDIS_HOST;
  }
  if (process.env.REDIS_PORT) {
    config.redis = config.redis || {};
    config.redis.port = parseInt(process.env.REDIS_PORT, 10);
  }

  // S3 overrides
  if (process.env.S3_ENDPOINT) {
    config.storage = config.storage || {};
    config.storage.s3 = config.storage.s3 || {};
    config.storage.s3.endpoint = process.env.S3_ENDPOINT;
  }
  if (process.env.S3_ACCESS_KEY) {
    config.storage = config.storage || {};
    config.storage.s3 = config.storage.s3 || {};
    config.storage.s3.access_key = process.env.S3_ACCESS_KEY;
  }
  if (process.env.S3_SECRET_KEY) {
    config.storage = config.storage || {};
    config.storage.s3 = config.storage.s3 || {};
    config.storage.s3.secret_key = process.env.S3_SECRET_KEY;
  }
  if (process.env.S3_BUCKET) {
    config.storage = config.storage || {};
    config.storage.s3 = config.storage.s3 || {};
    config.storage.s3.bucket = process.env.S3_BUCKET;
  }

  // JWT override
  if (process.env.JWT_SECRET) {
    config.security = config.security || {};
    config.security.jwt = config.security.jwt || {};
    config.security.jwt.secret = process.env.JWT_SECRET;
  }

  return config;
}

function validateConfig(config: any): void {
  const requiredFields = [
    ['app', 'name'],
    ['database', 'host'],
    ['database', 'name'],
    ['storage', 's3', 'bucket'],
  ];

  for (const fieldPath of requiredFields) {
    let current = config;
    for (const field of fieldPath) {
      if (current === undefined || current[field] === undefined) {
        throw new Error(`Missing required configuration field: ${fieldPath.join('.')}`);
      }
      current = current[field];
    }
  }
}

// Export singleton config instance
export const config = loadConfig();

// Helper function to get config
export function getConfig(): Config {
  return config;
}
