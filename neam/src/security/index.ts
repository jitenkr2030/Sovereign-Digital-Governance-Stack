/**
 * Security Module Index - Exports all security components
 * 
 * This file provides the main entry point for the National Dashboard
 * security module including HSM integration, cryptographic operations,
 * and certificate management.
 * 
 * @packageDocumentation
 */

// HSM Client
export { HSMClient } from './hsm/hsm-client';
export * from './hsm/hsm-client';

// Certificate Service
export { CertificateService } from './crypto/certificate-service';
export * from './crypto/certificate-service';

// TLS Manager
export { TLSCertificateManager } from './crypto/tls-manager';
export * from './crypto/tls-manager';

// Encryption Service
export { EncryptionService } from './crypto/encryption-service';
export * from './crypto/encryption-service';

// Signature Service
export { SignatureService } from './crypto/signature-service';
export * from './crypto/signature-service';

// Dashboard
export { CryptoDashboard } from './dashboard/CryptoDashboard';

// Security utilities
import * as crypto from 'crypto';

/**
 * Security module initialization helper
 */
export async function initializeSecurityModule(config: SecurityConfig): Promise<SecurityModule> {
  const hsmClient = new (await import('./hsm/hsm-client')).HSMClient({
    libraryPath: config.hsmLibraryPath,
    slotId: config.hsmSlotId,
    pin: config.hsmPin,
    softHSMFallback: config.enableSoftHSMFallback ?? true
  });

  await hsmClient.initialize();

  const certificateService = new (await import('./crypto/certificate-service')).CertificateService(hsmClient);
  await certificateService.initialize();

  const tlsManager = new (await import('./crypto/tls-manager')).TLSCertificateManager(
    certificateService,
    hsmClient
  );
  await tlsManager.initialize();

  const encryptionService = new (await import('./crypto/encryption-service')).EncryptionService(hsmClient);
  await encryptionService.initialize();

  const signatureService = new (await import('./crypto/signature-service')).SignatureService(hsmClient);
  await signatureService.initialize();

  return {
    hsmClient,
    certificateService,
    tlsManager,
    encryptionService,
    signatureService
  };
}

/**
 * Security configuration interface
 */
export interface SecurityConfig {
  hsmLibraryPath: string;
  hsmSlotId: number;
  hsmPin: string;
  enableSoftHSMFallback?: boolean;
  tlsConfig?: {
    port?: number;
    minVersion?: string;
    maxVersion?: string;
  };
  monitoringConfig?: {
    checkIntervalMinutes?: number;
    warningThresholdDays?: number;
    criticalThresholdDays?: number;
    autoRenewEnabled?: boolean;
  };
}

/**
 * Security module interface
 */
export interface SecurityModule {
  hsmClient: any;
  certificateService: any;
  tlsManager: any;
  encryptionService: any;
  signatureService: any;

  shutdown(): Promise<void>;
}

/**
 * Utility function to generate secure random bytes
 */
export function generateSecureRandom(length: number = 32): Buffer {
  return crypto.randomBytes(length);
}

/**
 * Utility function to generate secure random string
 */
export function generateSecureRandomString(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

/**
 * Utility function to hash data
 */
export function hashData(data: Buffer, algorithm: string = 'sha256'): Buffer {
  return crypto.createHash(algorithm).update(data).digest();
}

/**
 * Utility function to compare buffers in constant time
 */
export function secureCompare(a: Buffer, b: Buffer): boolean {
  if (a.length !== b.length) {
    return false;
  }
  return crypto.timingSafeEqual(a, b);
}

/**
 * Utility function to derive key from password
 */
export function deriveKeyFromPassword(
  password: string,
  salt: Buffer,
  iterations: number = 100000,
  keyLength: number = 32
): { key: Buffer; salt: Buffer } {
  const key = crypto.pbkdf2Sync(
    password,
    salt,
    iterations,
    keyLength,
    'sha256'
  );
  return { key, salt };
}

/**
 * Utility function to create password salt
 */
export function createPasswordSalt(): Buffer {
  return crypto.randomBytes(16);
}
