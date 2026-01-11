/**
 * Encryption Service - HSM-Based Data Encryption
 * 
 * This module provides data encryption and decryption services using
 * keys stored in Hardware Security Modules. It supports envelope encryption
 * where data encryption keys (DEKs) are wrapped by HSM master keys.
 * 
 * @packageDocumentation
 */

import * as crypto from 'crypto';
import { EventEmitter } from 'events';
import { v4 as uuidv4 } from 'uuid';
import { HSMClient, HSMKey, KeyType, KeyUsage, LogSeverity, SigningRequest, SigningMode } from '../hsm/hsm-client';

// Encryption types
export enum EncryptionAlgorithm {
  AES_256_GCM = 'AES-256-GCM',
  AES_256_CBC = 'AES-256-CBC',
  AES_128_GCM = 'AES-128-GCM',
  RSA_OAEP = 'RSA-OAEP',
  RSA_PKCS1 = 'RSA-PKCS1'
}

// Encryption mode
export enum EncryptionMode {
  ENCRYPT = 'ENCRYPT',
  DECRYPT = 'DECRYPT'
}

// Data encryption key (DEK)
export interface DataEncryptionKey {
  id: string;
  key: Buffer;
  algorithm: string;
  createdAt: Date;
  encryptedKey?: Buffer; // Wrapped by master key
  keyId?: string; // Reference to HSM key
}

// Encryption request
export interface EncryptionRequest {
  data: Buffer;
  algorithm: EncryptionAlgorithm;
  aad?: Buffer; // Additional Authenticated Data for GCM
  keyId?: string; // Use specific HSM key
  generateKey?: boolean; // Generate new DEK
}

// Decryption request
export interface DecryptionRequest {
  encryptedData: Buffer;
  iv: Buffer;
  tag?: Buffer; // Auth tag for GCM
  algorithm: EncryptionAlgorithm;
  keyId: string;
  aad?: Buffer;
}

// Encrypted result
export interface EncryptedResult {
  ciphertext: Buffer;
  iv: Buffer;
  tag?: Buffer;
  algorithm: EncryptionAlgorithm;
  keyId: string;
  encryptedKey?: Buffer; // Wrapped DEK
  keyId?: string; // Reference to HSM master key
}

// Key wrap request
export interface KeyWrapRequest {
  keyToWrap: Buffer;
  wrappingKeyId: string;
  algorithm: EncryptionAlgorithm;
}

// Key unwrap request
export interface KeyUnwrapRequest {
  wrappedKey: Buffer;
  unwrappingKeyId: string;
  algorithm: EncryptionAlgorithm;
}

// Encryption audit log entry
export interface EncryptionAuditEntry {
  id: string;
  timestamp: Date;
  operation: EncryptionMode;
  algorithm: string;
  keyId?: string;
  success: boolean;
  errorMessage?: string;
  dataSize: number;
  duration: number;
}

/**
 * Encryption Service - Handles all encryption operations
 */
export class EncryptionService extends EventEmitter {
  private hsmClient: HSMClient;
  private dekCache: Map<string, DataEncryptionKey> = new Map();
  private auditLogs: EncryptionAuditEntry[] = [];
  private masterKeyId: string | null = null;

  /**
   * Creates a new Encryption Service instance
   * @param hsmClient - HSM client for key operations
   */
  constructor(hsmClient: HSMClient) {
    super();
    this.hsmClient = hsmClient;
  }

  /**
   * Initializes the encryption service
   */
  async initialize(): Promise<void> {
    try {
      console.log('[ENCRYPTION-SERVICE] Initializing encryption service');

      // Initialize HSM client if not already done
      if (!this.hsmClient.getStatus().initialized) {
        await this.hsmClient.initialize();
      }

      // Find or create master key for key wrapping
      await this.ensureMasterKey();

      console.log('[ENCRYPTION-SERVICE] Encryption service initialized');
      console.log('[ENCRYPTION-SERVICE] Master key ID:', this.masterKeyId);

    } catch (error: any) {
      console.error('[ENCRYPTION-SERVICE] Initialization failed:', error.message);
      throw error;
    }
  }

  /**
   * Ensures a master key exists for key wrapping operations
   */
  private async ensureMasterKey(): Promise<void> {
    try {
      // Look for existing master key
      const keys = await this.hsmClient.listKeys();
      const masterKey = keys.find(k =>
        k.label.includes('MASTER_KEY') || k.label.includes('KEK')
      );

      if (masterKey) {
        this.masterKeyId = masterKey.id.toString('hex');
        console.log('[ENCRYPTION-SERVICE] Using existing master key:', this.masterKeyId);
      } else {
        // Generate new master key
        const newKey = await this.hsmClient.generateKey({
          label: `MASTER_KEY-${Date.now()}`,
          type: KeyType.AES,
          algorithm: 'AES-256-GCM',
          size: 256,
          extractable: false,
          usage: [KeyUsage.WRAP, KeyUsage.UNWRAP],
          validityPeriodDays: 365
        });
        this.masterKeyId = newKey.id.toString('hex');
        console.log('[ENCRYPTION-SERVICE] Created new master key:', this.masterKeyId);
      }

    } catch (error: any) {
      console.error('[ENCRYPTION-SERVICE] Failed to ensure master key:', error.message);
      throw error;
    }
  }

  /**
   * Encrypts data using envelope encryption
   * 
   * @param request - Encryption request parameters
   * @returns Encrypted result with wrapped key
   */
  async encrypt(request: EncryptionRequest): Promise<EncryptedResult> {
    const startTime = Date.now();

    try {
      console.log('[ENCRYPTION-SERVICE] Starting encryption');

      // Generate or use provided data encryption key (DEK)
      let dek: DataEncryptionKey;
      
      if (request.generateKey || !request.keyId) {
        dek = await this.generateDEK(request.algorithm);
      } else {
        dek = await this.getOrCreateDEK(request.keyId, request.algorithm);
      }

      // Encrypt data with DEK
      const result = await this.encryptWithDEK(request.data, dek, request.aad);

      // Wrap DEK with master key if using envelope encryption
      if (this.masterKeyId && !request.keyId) {
        const wrappedDEK = await this.wrapDEK(dek.key, this.masterKeyId);
        result.encryptedKey = wrappedDEK;
        result.keyId = this.masterKeyId;
      }

      // Update result with key information
      result.keyId = dek.id;
      result.algorithm = request.algorithm;

      const duration = Date.now() - startTime;

      // Log audit entry
      this.logAudit({
        operation: EncryptionMode.ENCRYPT,
        algorithm: request.algorithm,
        keyId: result.keyId,
        success: true,
        dataSize: request.data.length,
        duration
      });

      console.log('[ENCRYPTION-SERVICE] Encryption completed in', duration, 'ms');

      return result;

    } catch (error: any) {
      const duration = Date.now() - startTime;

      this.logAudit({
        operation: EncryptionMode.ENCRYPT,
        algorithm: request.algorithm,
        success: false,
        errorMessage: error.message,
        dataSize: request.data.length,
        duration
      });

      console.error('[ENCRYPTION-SERVICE] Encryption failed:', error.message);
      throw error;
    }
  }

  /**
   * Decrypts data using envelope encryption
   * 
   * @param request - Decryption request parameters
   * @returns Decrypted data
   */
  async decrypt(request: DecryptionRequest): Promise<Buffer> {
    const startTime = Date.now();

    try {
      console.log('[ENCRYPTION-SERVICE] Starting decryption');

      // Get the DEK (either from cache or unwrap)
      let dekKey: Buffer;

      if (request.encryptedKey) {
        // Unwrap DEK using master key
        dekKey = await this.unwrapDEK(
          request.encryptedKey,
          this.masterKeyId!
        );
      } else {
        // Get DEK from cache
        const dek = this.dekCache.get(request.keyId);
        if (!dek) {
          throw new Error(`DEK not found: ${request.keyId}`);
        }
        dekKey = dek.key;
      }

      // Decrypt data with DEK
      const decrypted = await this.decryptWithDEK(
        request.encryptedData,
        request.iv,
        dekKey,
        request.tag,
        request.algorithm,
        request.aad
      );

      const duration = Date.now() - startTime;

      // Log audit entry
      this.logAudit({
        operation: EncryptionMode.DECRYPT,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: true,
        dataSize: request.encryptedData.length,
        duration
      });

      console.log('[ENCRYPTION-SERVICE] Decryption completed in', duration, 'ms');

      return decrypted;

    } catch (error: any) {
      const duration = Date.now() - startTime;

      this.logAudit({
        operation: EncryptionMode.DECRYPT,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: false,
        errorMessage: error.message,
        dataSize: request.encryptedData.length,
        duration
      });

      console.error('[ENCRYPTION-SERVICE] Decryption failed:', error.message);
      throw error;
    }
  }

  /**
   * Generates a new data encryption key
   */
  private async generateDEK(algorithm: EncryptionAlgorithm): Promise<DataEncryptionKey> {
    const keySize = this.getKeySizeForAlgorithm(algorithm);
    const key = crypto.randomBytes(keySize);

    const dek: DataEncryptionKey = {
      id: uuidv4(),
      key,
      algorithm,
      createdAt: new Date()
    };

    // Cache the DEK
    this.dekCache.set(dek.id, dek);

    return dek;
  }

  /**
   * Gets or creates a DEK for a specific key
   */
  private async getOrCreateDEK(
    keyId: string,
    algorithm: EncryptionAlgorithm
  ): Promise<DataEncryptionKey> {
    // Check cache first
    let dek = this.dekCache.get(keyId);
    if (dek && dek.algorithm === algorithm) {
      return dek;
    }

    // Create new DEK
    dek = await this.generateDEK(algorithm);
    dek.keyId = keyId;
    this.dekCache.set(keyId, dek);

    return dek;
  }

  /**
   * Encrypts data using a DEK
   */
  private async encryptWithDEK(
    data: Buffer,
    dek: DataEncryptionKey,
    aad?: Buffer
  ): Promise<Partial<EncryptedResult>> {
    const iv = crypto.randomBytes(this.getIVSizeForAlgorithm(dek.algorithm));

    let ciphertext: Buffer;
    let tag: Buffer | undefined;

    switch (dek.algorithm) {
      case EncryptionAlgorithm.AES_256_GCM:
      case EncryptionAlgorithm.AES_128_GCM:
        const gcmCipher = crypto.createCipheriv(
          this.mapAlgorithm(dek.algorithm),
          dek.key,
          iv,
          { authTagLength: 16 }
        );
        if (aad) {
          gcmCipher.setAAD(aad);
        }
        ciphertext = Buffer.concat([
          gcmCipher.update(data),
          gcmCipher.final()
        ]);
        tag = gcmCipher.getAuthTag();
        break;

      case EncryptionAlgorithm.AES_256_CBC:
        const cbcCipher = crypto.createCipheriv(
          this.mapAlgorithm(dek.algorithm),
          dek.key,
          iv
        );
        ciphertext = Buffer.concat([
          cbcCipher.update(data),
          cbcCipher.final()
        ]);
        break;

      default:
        throw new Error(`Unsupported algorithm: ${dek.algorithm}`);
    }

    return {
      ciphertext,
      iv,
      tag
    };
  }

  /**
   * Decrypts data using a DEK
   */
  private async decryptWithDEK(
    ciphertext: Buffer,
    iv: Buffer,
    dekKey: Buffer,
    tag: Buffer | undefined,
    algorithm: EncryptionAlgorithm,
    aad?: Buffer
  ): Promise<Buffer> {
    let decrypted: Buffer;

    switch (algorithm) {
      case EncryptionAlgorithm.AES_256_GCM:
      case EncryptionAlgorithm.AES_128_GCM:
        const gcmDecipher = crypto.createDecipheriv(
          this.mapAlgorithm(algorithm),
          dekKey,
          iv,
          { authTagLength: 16 }
        );
        if (aad) {
          gcmDecipher.setAAD(aad);
        }
        if (tag) {
          gcmDecipher.setAuthTag(tag);
        }
        decrypted = Buffer.concat([
          gcmDecipher.update(ciphertext),
          gcmDecipher.final()
        ]);
        break;

      case EncryptionAlgorithm.AES_256_CBC:
        const cbcDecipher = crypto.createDecipheriv(
          this.mapAlgorithm(algorithm),
          dekKey,
          iv
        );
        decrypted = Buffer.concat([
          cbcDecipher.update(ciphertext),
          cbcDecipher.final()
        ]);
        break;

      default:
        throw new Error(`Unsupported algorithm: ${algorithm}`);
    }

    return decrypted;
  }

  /**
   * Wraps (encrypts) a DEK using the master key in HSM
   */
  private async wrapDEK(dekKey: Buffer, wrappingKeyId: string): Promise<Buffer> {
    // Use HSM for key wrapping
    const wrappedKey = await this.hsmClient.unwrapKey({
      wrappedKey: dekKey,
      wrappingKeyId,
      algorithm: 'AES-256-GCM'
    });

    // The HSM returns the wrapped key - in production this would stay in HSM
    return wrappedKey.id;
  }

  /**
   * Unwraps a DEK using the master key in HSM
   */
  private async unwrapDEK(wrappedKey: Buffer, unwrappingKeyId: string): Promise<Buffer> {
    // Use HSM for key unwrapping
    const unwrappedKey = await this.hsmClient.unwrapKey({
      wrappedKey,
      wrappingKeyId: unwrappingKeyId,
      algorithm: 'AES-256-GCM'
    });

    // Return the key handle (not actual key material)
    return unwrappedKey.id;
  }

  /**
   * Directly encrypts data using HSM key (for small data)
   */
  async encryptWithHSMKey(
    data: Buffer,
    keyId: string,
    algorithm: EncryptionAlgorithm = EncryptionAlgorithm.RSA_OAEP
  ): Promise<EncryptedResult> {
    const startTime = Date.now();

    try {
      console.log('[ENCRYPTION-SERVICE] Direct HSM encryption');

      // Encrypt with HSM key
      const result = await this.hsmClient.sign({
        keyId,
        data,
        algorithm: 'RSA-OAEP',
        mode: SigningMode.DIRECT
      });

      const duration = Date.now() - startTime;

      this.logAudit({
        operation: EncryptionMode.ENCRYPT,
        algorithm,
        keyId,
        success: true,
        dataSize: data.length,
        duration
      });

      return {
        ciphertext: result.signature,
        iv: Buffer.from(''),
        algorithm,
        keyId
      };

    } catch (error: any) {
      console.error('[ENCRYPTION-SERVICE] HSM encryption failed:', error.message);
      throw error;
    }
  }

  /**
   * Directly decrypts data using HSM key (for small data)
   */
  async decryptWithHSMKey(
    encryptedData: Buffer,
    keyId: string
  ): Promise<Buffer> {
    // RSA decryption using HSM would be similar to sign operation
    // In PKCS#11, C_Decrypt is used
    throw new Error('HSM decryption not yet implemented - use envelope encryption');
  }

  /**
   * Derives a key from a password using PBKDF2
   */
  deriveKeyFromPassword(
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
   * Encrypts a string value (base64 encoding for storage)
   */
  async encryptString(
    plaintext: string,
    algorithm: EncryptionAlgorithm = EncryptionAlgorithm.AES_256_GCM
  ): Promise<string> {
    const data = Buffer.from(plaintext, 'utf8');
    const result = await this.encrypt({
      data,
      algorithm,
      generateKey: true
    });

    // Combine components for storage
    const combined = Buffer.concat([
      result.iv,
      result.tag || Buffer.from(''),
      Buffer.from([0]), // separator
      result.ciphertext
    ]);

    return combined.toString('base64');
  }

  /**
   * Decrypts a string value
   */
  async decryptString(
    encryptedBase64: string,
    algorithm: EncryptionAlgorithm = EncryptionAlgorithm.AES_256_GCM
  ): Promise<string> {
    const combined = Buffer.from(encryptedBase64, 'base64');
    
    // Parse components (simplified - in production use proper parsing)
    const parts = this.parseEncryptedPackage(combined);

    const result = await this.decrypt({
      encryptedData: parts.ciphertext,
      iv: parts.iv,
      tag: parts.tag,
      algorithm,
      keyId: parts.keyId || this.masterKeyId!
    });

    return result.toString('utf8');
  }

  /**
   * Parses encrypted package into components
   */
  private parseEncryptedPackage(
    combined: Buffer
  ): { iv: Buffer; tag: Buffer | undefined; ciphertext: Buffer; keyId?: string } {
    // Simplified parsing - in production use proper format
    const ivLength = 12; // GCM IV size
    const tagLength = 16; // GCM auth tag size

    const iv = combined.subarray(0, ivLength);
    const tag = combined.subarray(ivLength, ivLength + tagLength);
    const separatorIndex = combined.indexOf(Buffer.from([0]), ivLength + tagLength);
    const ciphertext = combined.subarray(separatorIndex + 1);

    return { iv, tag: tag.length > 0 ? tag : undefined, ciphertext };
  }

  /**
   * Generates a secure random string
   */
  generateSecureRandom(length: number = 32): string {
    return crypto.randomBytes(length).toString('hex');
  }

  /**
   * Clears the DEK cache
   */
  clearCache(): void {
    this.dekCache.clear();
    console.log('[ENCRYPTION-SERVICE] DEK cache cleared');
  }

  /**
   * Gets audit logs
   */
  getAuditLogs(): EncryptionAuditEntry[] {
    return [...this.auditLogs];
  }

  /**
   * Clears audit logs
   */
  clearAuditLogs(): void {
    this.auditLogs = [];
  }

  /**
   * Gets service status
   */
  getStatus(): {
    initialized: boolean;
    masterKeyId: string | null;
    cachedDEKs: number;
    auditLogCount: number;
  } {
    return {
      initialized: this.hsmClient.getStatus().initialized,
      masterKeyId: this.masterKeyId,
      cachedDEKs: this.dekCache.size,
      auditLogCount: this.auditLogs.length
    };
  }

  // Private helper methods

  /**
   * Logs an audit entry
   */
  private logAudit(entry: Omit<EncryptionAuditEntry, 'id' | 'timestamp'>): void {
    const fullEntry: EncryptionAuditEntry = {
      ...entry,
      id: uuidv4(),
      timestamp: new Date()
    };

    this.auditLogs.push(fullEntry);
    this.emit('audit', fullEntry);

    // Console logging for debugging
    const timestamp = fullEntry.timestamp.toISOString();
    console.log(`[ENCRYPTION-AUDIT] ${timestamp} [${fullEntry.operation}] ${fullEntry.algorithm}: ${fullEntry.success ? 'SUCCESS' : 'FAILURE'}`);
  }

  /**
   * Maps algorithm enum to crypto module algorithm
   */
  private mapAlgorithm(algorithm: EncryptionAlgorithm): string {
    const map: Record<EncryptionAlgorithm, string> = {
      [EncryptionAlgorithm.AES_256_GCM]: 'aes-256-gcm',
      [EncryptionAlgorithm.AES_256_CBC]: 'aes-256-cbc',
      [EncryptionAlgorithm.AES_128_GCM]: 'aes-128-gcm',
      [EncryptionAlgorithm.RSA_OAEP]: 'rsa',
      [EncryptionAlgorithm.RSA_PKCS1]: 'rsa'
    };

    return map[algorithm] || 'aes-256-gcm';
  }

  /**
   * Gets key size for algorithm
   */
  private getKeySizeForAlgorithm(algorithm: EncryptionAlgorithm): number {
    const sizes: Record<EncryptionAlgorithm, number> = {
      [EncryptionAlgorithm.AES_256_GCM]: 32,
      [EncryptionAlgorithm.AES_256_CBC]: 32,
      [EncryptionAlgorithm.AES_128_GCM]: 16,
      [EncryptionAlgorithm.RSA_OAEP]: 256,
      [EncryptionAlgorithm.RSA_PKCS1]: 256
    };

    return sizes[algorithm] || 32;
  }

  /**
   * Gets IV size for algorithm
   */
  private getIVSizeForAlgorithm(algorithm: EncryptionAlgorithm): number {
    const sizes: Record<EncryptionAlgorithm, number> = {
      [EncryptionAlgorithm.AES_256_GCM]: 12,
      [EncryptionAlgorithm.AES_256_CBC]: 16,
      [EncryptionAlgorithm.AES_128_GCM]: 12,
      [EncryptionAlgorithm.RSA_OAEP]: 0,
      [EncryptionAlgorithm.RSA_PKCS1]: 0
    };

    return sizes[algorithm] || 12;
  }

  /**
   * Shuts down the encryption service
   */
  async shutdown(): Promise<void> {
    console.log('[ENCRYPTION-SERVICE] Shutting down');
    this.clearCache();
    this.clearAuditLogs();
    console.log('[ENCRYPTION-SERVICE] Encryption service shutdown complete');
  }
}

// Export types
export {
  EncryptionAlgorithm,
  EncryptionMode,
  DataEncryptionKey,
  EncryptionRequest,
  DecryptionRequest,
  EncryptedResult,
  KeyWrapRequest,
  KeyUnwrapRequest,
  EncryptionAuditEntry
};
