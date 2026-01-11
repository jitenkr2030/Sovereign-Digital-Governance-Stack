/**
 * Signature Service - Digital Signature Verification and Management
 * 
 * This module provides digital signature creation and verification services
 * using keys stored in Hardware Security Modules. It supports multiple
 * signature algorithms and provides comprehensive audit logging.
 * 
 * @packageDocumentation
 */

import * as crypto from 'crypto';
import { EventEmitter } from 'events';
import { v4 as uuidv4 } from 'uuid';
import { HSMClient, HSMKey, KeyType, KeyUsage, LogSeverity, SigningRequest, SigningMode } from '../hsm/hsm-client';

// Signature algorithms
export enum SignatureAlgorithm {
  RSA_SHA256 = 'RSA-SHA256',
  RSA_SHA384 = 'RSA-SHA384',
  RSA_SHA512 = 'RSA-SHA512',
  ECDSA_SHA256 = 'ECDSA-SHA256',
  ECDSA_SHA384 = 'ECDSA-SHA384',
  ECDSA_SHA512 = 'ECDSA-SHA512',
  RSASSA_PSS = 'RSASSA-PSS',
  HMAC_SHA256 = 'HMAC-SHA256'
}

// Signature mode
export enum SignatureMode {
  SIGN = 'SIGN',
  VERIFY = 'VERIFY'
}

// Signature request
export interface SignatureRequest {
  data: Buffer;
  keyId: string;
  algorithm: SignatureAlgorithm;
  mode?: SignatureMode;
  padding?: 'pkcs1' | 'pss';
}

// Signature result
export interface SignatureResult {
  signature: Buffer;
  keyId: string;
  algorithm: SignatureAlgorithm;
  timestamp: Date;
  duration: number;
}

// Verification request
export interface VerificationRequest {
  data: Buffer;
  signature: Buffer;
  keyId: string;
  algorithm: SignatureAlgorithm;
  publicKey?: Buffer;
}

// Verification result
export interface VerificationResult {
  valid: boolean;
  keyId: string;
  algorithm: SignatureAlgorithm;
  timestamp: Date;
  error?: string;
}

// Signature policy
export interface SignaturePolicy {
  id: string;
  name: string;
  allowedAlgorithms: SignatureAlgorithm[];
  requireTimestamp: boolean;
  maxDataSize: number;
  validityPeriodSeconds?: number;
}

// Batch signature request
export interface BatchSignatureRequest {
  items: Array<{
    data: Buffer;
    metadata?: Record<string, any>;
  }>;
  keyId: string;
  algorithm: SignatureAlgorithm;
}

// Batch signature result
export interface BatchSignatureResult {
  signatures: Array<{
    signature: string;
    index: number;
    metadata?: Record<string, any>;
    error?: string;
  }>;
  totalProcessed: number;
  totalFailed: number;
  duration: number;
}

// Signature audit entry
export interface SignatureAuditEntry {
  id: string;
  timestamp: Date;
  operation: SignatureMode;
  algorithm: string;
  keyId: string;
  keyLabel?: string;
  success: boolean;
  errorMessage?: string;
  dataSize: number;
  signatureSize: number;
  duration: number;
}

/**
 * Signature Service - Handles all signature operations
 */
export class SignatureService extends EventEmitter {
  private hsmClient: HSMClient;
  private auditLogs: SignatureAuditEntry[] = [];
  private policyCache: Map<string, SignaturePolicy> = new Map();
  private defaultPolicy: SignaturePolicy;

  /**
   * Creates a new Signature Service instance
   * @param hsmClient - HSM client for cryptographic operations
   */
  constructor(hsmClient: HSMClient) {
    super();
    this.hsmClient = hsmClient;

    // Initialize default policy
    this.defaultPolicy = {
      id: 'default',
      name: 'Default Signature Policy',
      allowedAlgorithms: [
        SignatureAlgorithm.RSA_SHA256,
        SignatureAlgorithm.RSA_SHA384,
        SignatureAlgorithm.RSA_SHA512,
        SignatureAlgorithm.ECDSA_SHA256,
        SignatureAlgorithm.HMAC_SHA256
      ],
      requireTimestamp: false,
      maxDataSize: 1024 * 1024 // 1MB
    };

    // Register default policy
    this.policyCache.set('default', this.defaultPolicy);
  }

  /**
   * Initializes the signature service
   */
  async initialize(): Promise<void> {
    try {
      console.log('[SIGNATURE-SERVICE] Initializing signature service');

      // Initialize HSM client if not already done
      if (!this.hsmClient.getStatus().initialized) {
        await this.hsmClient.initialize();
      }

      console.log('[SIGNATURE-SERVICE] Signature service initialized');
    } catch (error: any) {
      console.error('[SIGNATURE-SERVICE] Initialization failed:', error.message);
      throw error;
    }
  }

  /**
   * Creates a digital signature
   * 
   * @param request - Signature request parameters
   * @returns Signature result
   */
  async sign(request: SignatureRequest): Promise<SignatureResult> {
    const startTime = Date.now();

    try {
      console.log('[SIGNATURE-SERVICE] Creating signature with key:', request.keyId);

      // Validate policy
      const policy = this.defaultPolicy;
      if (!policy.allowedAlgorithms.includes(request.algorithm)) {
        throw new Error(`Algorithm ${request.algorithm} not allowed by policy`);
      }

      // Check data size
      if (request.data.length > policy.maxDataSize) {
        throw new Error(`Data size ${request.data.length} exceeds maximum ${policy.maxDataSize}`);
      }

      // Create signing request for HSM
      const hsmRequest: SigningRequest = {
        keyId: request.keyId,
        data: request.data,
        algorithm: this.mapAlgorithmToHSM(request.algorithm),
        mode: SigningMode.HASH // HSM will handle hashing
      };

      // Sign with HSM
      const hsmResult = await this.hsmClient.sign(hsmRequest);

      const duration = Date.now() - startTime;

      // Log audit entry
      this.logAudit({
        operation: SignatureMode.SIGN,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: true,
        dataSize: request.data.length,
        signatureSize: hsmResult.signature.length,
        duration
      });

      console.log('[SIGNATURE-SERVICE] Signature created in', duration, 'ms');

      return {
        signature: hsmResult.signature,
        keyId: request.keyId,
        algorithm: request.algorithm,
        timestamp: hsmResult.timestamp,
        duration
      };

    } catch (error: any) {
      const duration = Date.now() - startTime;

      this.logAudit({
        operation: SignatureMode.SIGN,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: false,
        errorMessage: error.message,
        dataSize: request.data.length,
        signatureSize: 0,
        duration
      });

      console.error('[SIGNATURE-SERVICE] Signature failed:', error.message);
      throw error;
    }
  }

  /**
   * Verifies a digital signature
   * 
   * @param request - Verification request parameters
   * @returns Verification result
   */
  async verify(request: VerificationRequest): Promise<VerificationResult> {
    const startTime = Date.now();

    try {
      console.log('[SIGNATURE-SERVICE] Verifying signature with key:', request.keyId);

      // Validate policy
      const policy = this.defaultPolicy;
      if (!policy.allowedAlgorithms.includes(request.algorithm)) {
        throw new Error(`Algorithm ${request.algorithm} not allowed by policy`);
      }

      // Verify using HSM
      const isValid = await this.hsmClient.verify(
        request.keyId,
        request.data,
        request.signature,
        this.mapAlgorithmToHSM(request.algorithm)
      );

      const duration = Date.now() - startTime;

      // Log audit entry
      this.logAudit({
        operation: SignatureMode.VERIFY,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: true,
        dataSize: request.data.length,
        signatureSize: request.signature.length,
        duration
      });

      console.log('[SIGNATURE-SERVICE] Verification completed in', duration, 'ms, valid:', isValid);

      return {
        valid: isValid,
        keyId: request.keyId,
        algorithm: request.algorithm,
        timestamp: new Date()
      };

    } catch (error: any) {
      const duration = Date.now() - startTime;

      this.logAudit({
        operation: SignatureMode.VERIFY,
        algorithm: request.algorithm,
        keyId: request.keyId,
        success: false,
        errorMessage: error.message,
        dataSize: request.data.length,
        signatureSize: request.signature.length,
        duration
      });

      console.error('[SIGNATURE-SERVICE] Verification failed:', error.message);

      return {
        valid: false,
        keyId: request.keyId,
        algorithm: request.algorithm,
        timestamp: new Date(),
        error: error.message
      };
    }
  }

  /**
   * Signs data with HMAC (symmetric)
   */
  async signHMAC(
    data: Buffer,
    key: Buffer,
    algorithm: SignatureAlgorithm = SignatureAlgorithm.HMAC_SHA256
  ): Promise<Buffer> {
    const startTime = Date.now();

    try {
      console.log('[SIGNATURE-SERVICE] Creating HMAC signature');

      const hmac = crypto.createHmac(
        this.mapHMACAlgorithm(algorithm),
        key
      );
      hmac.update(data);
      const signature = hmac.digest();

      const duration = Date.now() - startTime;

      this.logAudit({
        operation: SignatureMode.SIGN,
        algorithm: algorithm.toString(),
        keyId: 'hmac-internal',
        success: true,
        dataSize: data.length,
        signatureSize: signature.length,
        duration
      });

      return signature;

    } catch (error: any) {
      console.error('[SIGNATURE-SERVICE] HMAC signing failed:', error.message);
      throw error;
    }
  }

  /**
   * Verifies HMAC signature
   */
  async verifyHMAC(
    data: Buffer,
    signature: Buffer,
    key: Buffer,
    algorithm: SignatureAlgorithm = SignatureAlgorithm.HMAC_SHA256
  ): Promise<boolean> {
    const startTime = Date.now();

    try {
      const expectedSignature = await this.signHMAC(data, key, algorithm);
      const isValid = crypto.timingSafeEqual(expectedSignature, signature);

      const duration = Date.now() - startTime;

      this.logAudit({
        operation: SignatureMode.VERIFY,
        algorithm: algorithm.toString(),
        keyId: 'hmac-internal',
        success: true,
        dataSize: data.length,
        signatureSize: signature.length,
        duration
      });

      return isValid;

    } catch (error: any) {
      console.error('[SIGNATURE-SERVICE] HMAC verification failed:', error.message);
      return false;
    }
  }

  /**
   * Creates batch signatures for multiple items
   * 
   * @param request - Batch signature request
   * @returns Batch signature result
   */
  async signBatch(request: BatchSignatureRequest): Promise<BatchSignatureResult> {
    const startTime = Date.now();
    const results: BatchSignatureResult = {
      signatures: [],
      totalProcessed: 0,
      totalFailed: 0,
      duration: 0
    };

    console.log('[SIGNATURE-SERVICE] Processing batch of', request.items.length, 'items');

    for (let i = 0; i < request.items.length; i++) {
      const item = request.items[i];
      
      try {
        const result = await this.sign({
          data: item.data,
          keyId: request.keyId,
          algorithm: request.algorithm
        });

        results.signatures.push({
          signature: result.signature.toString('base64'),
          index: i,
          metadata: item.metadata
        });
        results.totalProcessed++;

      } catch (error: any) {
        results.signatures.push({
          signature: '',
          index: i,
          metadata: item.metadata,
          error: error.message
        });
        results.totalFailed++;
      }
    }

    results.duration = Date.now() - startTime;

    console.log('[SIGNATURE-SERVICE] Batch completed:', 
      results.totalProcessed, 'success,', results.totalFailed, 'failed');

    return results;
  }

  /**
   * Verifies batch signatures
   */
  async verifyBatch(
    items: Array<{
      data: Buffer;
      signature: Buffer;
      keyId: string;
      algorithm: SignatureAlgorithm;
    }>
  ): Promise<{
    results: Array<VerificationResult>;
    totalValid: number;
    totalInvalid: number;
  }> {
    const results: Array<VerificationResult> = [];
    let totalValid = 0;
    let totalInvalid = 0;

    for (const item of items) {
      const result = await this.verify({
        data: item.data,
        signature: item.signature,
        keyId: item.keyId,
        algorithm: item.algorithm
      });

      results.push(result);
      if (result.valid) {
        totalValid++;
      } else {
        totalInvalid++;
      }
    }

    return { results, totalValid, totalInvalid };
  }

  /**
   * Creates a detached signature (signature separate from data)
   */
  async createDetachedSignature(
    data: Buffer,
    keyId: string,
    algorithm: SignatureAlgorithm = SignatureAlgorithm.RSA_SHA256
  ): Promise<string> {
    const result = await this.sign({
      data,
      keyId,
      algorithm
    });

    // Create detached signature package
    const signaturePackage = {
      algorithm: result.algorithm,
      keyId: result.keyId,
      timestamp: result.timestamp.toISOString(),
      signature: result.signature.toString('base64')
    };

    return Buffer.from(JSON.stringify(signaturePackage)).toString('base64');
  }

  /**
   * Verifies a detached signature
   */
  async verifyDetachedSignature(
    data: Buffer,
    detachedSignature: string,
    keyId: string
  ): Promise<VerificationResult> {
    try {
      const signaturePackage = JSON.parse(
        Buffer.from(detachedSignature, 'base64').toString('utf8')
      );

      const signature = Buffer.from(signaturePackage.signature, 'base64');

      return await this.verify({
        data,
        signature,
        keyId,
        algorithm: signaturePackage.algorithm
      });

    } catch (error: any) {
      return {
        valid: false,
        keyId,
        algorithm: SignatureAlgorithm.RSA_SHA256,
        timestamp: new Date(),
        error: `Invalid signature package: ${error.message}`
      };
    }
  }

  /**
   * Generates a signature fingerprint
   */
  generateFingerprint(signature: Buffer): string {
    const hash = crypto.createHash('sha256').update(signature).digest();
    return hash.subarray(0, 8).toString('hex').toUpperCase();
  }

  /**
   * Adds a signature policy
   */
  addPolicy(policy: SignaturePolicy): void {
    this.policyCache.set(policy.id, policy);
    console.log('[SIGNATURE-SERVICE] Policy added:', policy.name);
  }

  /**
   * Gets a signature policy
   */
  getPolicy(policyId: string): SignaturePolicy | undefined {
    return this.policyCache.get(policyId);
  }

  /**
   * Lists all policies
   */
  listPolicies(): SignaturePolicy[] {
    return Array.from(this.policyCache.values());
  }

  /**
   * Signs a hash directly (pre-hashed data)
   */
  async signHash(
    hash: Buffer,
    keyId: string,
    algorithm: SignatureAlgorithm
  ): Promise<SignatureResult> {
    return this.sign({
      data: hash,
      keyId,
      algorithm,
      mode: SignatureMode.SIGN
    });
  }

  /**
   * Verifies a hash signature
   */
  async verifyHash(
    hash: Buffer,
    signature: Buffer,
    keyId: string,
    algorithm: SignatureAlgorithm
  ): Promise<VerificationResult> {
    return this.verify({
      data: hash,
      signature,
      keyId,
      algorithm
    });
  }

  /**
   * Gets signature statistics
   */
  getStatistics(): {
    totalSignatures: number;
    totalVerifications: number;
    successRate: number;
    averageSignDuration: number;
    averageVerifyDuration: number;
  } {
    const signLogs = this.auditLogs.filter(l => l.operation === SignatureMode.SIGN);
    const verifyLogs = this.auditLogs.filter(l => l.operation === SignatureMode.VERIFY);

    const successfulSigns = signLogs.filter(l => l.success).length;
    const successfulVerifies = verifyLogs.filter(l => l.success).length;

    const avgSignDuration = signLogs.length > 0
      ? signLogs.reduce((sum, l) => sum + l.duration, 0) / signLogs.length
      : 0;

    const avgVerifyDuration = verifyLogs.length > 0
      ? verifyLogs.reduce((sum, l) => sum + l.duration, 0) / verifyLogs.length
      : 0;

    return {
      totalSignatures: signLogs.length,
      totalVerifications: verifyLogs.length,
      successRate: (signLogs.length + verifyLogs.length) > 0
        ? ((successfulSigns + successfulVerifies) / (signLogs.length + verifyLogs.length)) * 100
        : 100,
      averageSignDuration: Math.round(avgSignDuration),
      averageVerifyDuration: Math.round(avgVerifyDuration)
    };
  }

  /**
   * Gets audit logs
   */
  getAuditLogs(): SignatureAuditEntry[] {
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
    policyCount: number;
    auditLogCount: number;
  } {
    return {
      initialized: this.hsmClient.getStatus().initialized,
      policyCount: this.policyCache.size,
      auditLogCount: this.auditLogs.length
    };
  }

  // Private helper methods

  /**
   * Logs an audit entry
   */
  private logAudit(
    entry: Omit<SignatureAuditEntry, 'id' | 'timestamp'>
  ): void {
    const fullEntry: SignatureAuditEntry = {
      ...entry,
      id: uuidv4(),
      timestamp: new Date()
    };

    this.auditLogs.push(fullEntry);
    this.emit('audit', fullEntry);

    // Console logging
    const timestamp = fullEntry.timestamp.toISOString();
    console.log(`[SIGNATURE-AUDIT] ${timestamp} [${fullEntry.operation}] ${fullEntry.algorithm}: ${fullEntry.success ? 'SUCCESS' : 'FAILURE'}`);
  }

  /**
   * Maps algorithm to HSM format
   */
  private mapAlgorithmToHSM(algorithm: SignatureAlgorithm): string {
    const map: Record<SignatureAlgorithm, string> = {
      [SignatureAlgorithm.RSA_SHA256]: 'RSA-SHA256',
      [SignatureAlgorithm.RSA_SHA384]: 'RSA-SHA384',
      [SignatureAlgorithm.RSA_SHA512]: 'RSA-SHA512',
      [SignatureAlgorithm.ECDSA_SHA256]: 'ECDSA-SHA256',
      [SignatureAlgorithm.ECDSA_SHA384]: 'ECDSA-SHA384',
      [SignatureAlgorithm.ECDSA_SHA512]: 'ECDSA-SHA512',
      [SignatureAlgorithm.RSASSA_PSS]: 'RSASSA-PSS',
      [SignatureAlgorithm.HMAC_SHA256]: 'HMAC-SHA256'
    };

    return map[algorithm] || 'RSA-SHA256';
  }

  /**
   * Maps HMAC algorithm
   */
  private mapHMACAlgorithm(algorithm: SignatureAlgorithm): string {
    const map: Record<SignatureAlgorithm, string> = {
      [SignatureAlgorithm.RSA_SHA256]: 'sha256',
      [SignatureAlgorithm.RSA_SHA384]: 'sha384',
      [SignatureAlgorithm.RSA_SHA512]: 'sha512',
      [SignatureAlgorithm.ECDSA_SHA256]: 'sha256',
      [SignatureAlgorithm.ECDSA_SHA384]: 'sha384',
      [SignatureAlgorithm.ECDSA_SHA512]: 'sha512',
      [SignatureAlgorithm.RSASSA_PSS]: 'sha256',
      [SignatureAlgorithm.HMAC_SHA256]: 'sha256'
    };

    return map[algorithm] || 'sha256';
  }

  /**
   * Shuts down the signature service
   */
  async shutdown(): Promise<void> {
    console.log('[SIGNATURE-SERVICE] Shutting down');
    this.clearAuditLogs();
    console.log('[SIGNATURE-SERVICE] Signature service shutdown complete');
  }
}

// Export types
export {
  SignatureAlgorithm,
  SignatureMode,
  SignatureRequest,
  SignatureResult,
  VerificationRequest,
  VerificationResult,
  SignaturePolicy,
  BatchSignatureRequest,
  BatchSignatureResult,
  SignatureAuditEntry
};
