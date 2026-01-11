/**
 * PKCS#11 Client Wrapper for HSM Integration
 * 
 * This module provides a unified interface for interacting with Hardware Security Modules
 * through the PKCS#11 standard. It supports both physical HSMs and SoftHSM for development.
 * 
 * @packageDocumentation
 */

import * as pkcs11 from 'pkcs11js';
import * as crypto from 'crypto';
import { EventEmitter } from 'events';
import { v4 as uuidv4 } from 'uuid';

// Type definitions for PKCS11 module
interface PKCS11Lib {
  load(path: string): void;
  C_GetFunctionList(): any;
}

interface HSMKey {
  id: Buffer;
  label: string;
  type: KeyType;
  algorithm: string;
  size: number;
  extractable: boolean;
  sensitive: boolean;
  createdAt: Date;
  expiresAt?: Date;
  status: KeyStatus;
  usage: KeyUsage[];
}

interface Certificate {
  id: Buffer;
  label: string;
  subject: string;
  issuer: string;
  serialNumber: string;
  notBefore: Date;
  notAfter: Date;
  publicKey: Buffer;
  algorithm: string;
  status: CertificateStatus;
}

interface SigningRequest {
  keyId: string;
  data: Buffer;
  algorithm: string;
  mode: SigningMode;
}

interface SigningResult {
  signature: Buffer;
  keyId: string;
  timestamp: Date;
  algorithm: string;
}

interface HSMConfig {
  libraryPath: string;
  slotId: number;
  pin: string;
  readOnly?: boolean;
  softHSMFallback?: boolean;
}

interface KeyGenerationRequest {
  label: string;
  type: KeyType;
  algorithm: string;
  size: number;
  extractable?: boolean;
  usage?: KeyUsage[];
  validityPeriodDays?: number;
}

interface KeyUnwrapRequest {
  wrappedKey: Buffer;
  wrappingKeyId: string;
  algorithm: string;
}

enum KeyType {
  RSA = 'RSA',
  ECDSA = 'ECDSA',
  AES = 'AES',
  HMAC = 'HMAC'
}

enum KeyStatus {
  ACTIVE = 'ACTIVE',
  ROTATING = 'ROTATING',
  RETIRED = 'RETIRED',
  COMPROMISED = 'COMPROMISED'
}

enum KeyUsage {
  SIGN = 'SIGN',
  VERIFY = 'VERIFY',
  ENCRYPT = 'ENCRYPT',
  DECRYPT = 'DECRYPT',
  WRAP = 'WRAP',
  UNWRAP = 'UNWRAP'
}

enum CertificateStatus {
  VALID = 'VALID',
  EXPIRED = 'EXPIRED',
  REVOKED = 'REVOKED',
  PENDING = 'PENDING'
}

enum SigningMode {
  DIRECT = 'DIRECT',
  HASH = 'HASH',
  PSS = 'PSS'
}

enum LogSeverity {
  INFO = 'INFO',
  WARNING = 'WARNING',
  ERROR = 'ERROR',
  CRITICAL = 'CRITICAL'
}

interface AuditLogEntry {
  id: string;
  timestamp: Date;
  operation: string;
  keyId?: string;
  keyLabel?: string;
  slotId: number;
  result: 'SUCCESS' | 'FAILURE';
  errorMessage?: string;
  actorId: string;
  severity: LogSeverity;
  duration: number;
  metadata?: Record<string, any>;
}

/**
 * HSM Client - Main class for HSM interactions
 * 
 * Provides a secure interface for cryptographic operations with Hardware Security Modules.
 * Implements automatic reconnection, key lifecycle management, and comprehensive audit logging.
 * 
 * @example
 * ```typescript
 * const hsmClient = new HSMClient({
 *   libraryPath: '/usr/lib/softhsm/libsofthsm2.so',
 *   slotId: 0,
 *   pin: '123456'
 * });
 * 
 * await hsmClient.initialize();
 * const keys = await hsmClient.listKeys();
 * ```
 */
export class HSMClient extends EventEmitter {
  private pkcs11: any;
  private lib: PKCS11Lib | null = null;
  private config: HSMConfig;
  private isInitialized: boolean = false;
  private session: any = null;
  private slotInfo: any = null;
  private connectionRetryCount: number = 0;
  private maxRetries: number = 3;
  private reconnectDelay: number = 5000;
  private auditLogs: AuditLogEntry[] = [];
  private keyCache: Map<string, HSMKey> = new Map();
  private certificateCache: Map<string, Certificate> = new Map();

  /**
   * Creates a new HSM Client instance
   * @param config - HSM connection configuration
   */
  constructor(config: HSMConfig) {
    super();
    this.config = {
      readOnly: false,
      softHSMFallback: true,
      ...config
    };
  }

  /**
   * Initializes the HSM connection and establishes a session
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      this.logAudit('INITIALIZE', 'HSM already initialized', 'INFO');
      return;
    }

    try {
      this.logAudit('INITIALIZE', 'Starting HSM initialization', 'INFO');

      // Load PKCS#11 library
      this.loadLibrary();

      // Open session to HSM slot
      await this.openSession();

      // Login to HSM
      await this.login();

      // Cache existing keys and certificates
      await this.refreshCache();

      this.isInitialized = true;
      this.connectionRetryCount = 0;
      
      this.logAudit('INITIALIZE', 'HSM initialization successful', 'INFO');
      this.emit('initialized', { slotId: this.config.slotId });

    } catch (error: any) {
      this.logAudit('INITIALIZE', `HSM initialization failed: ${error.message}`, 'CRITICAL');
      await this.handleInitializationError(error);
      throw error;
    }
  }

  /**
   * Loads the PKCS#11 library from the specified path
   */
  private loadLibrary(): void {
    try {
      this.logAudit('LOAD_LIBRARY', `Loading library: ${this.config.libraryPath}`, 'INFO');

      // Dynamic import of PKCS#11 module
      this.pkcs11 = new pkcs11.PKCS11();
      this.pkcs11.load(this.config.libraryPath);

      // Get function list
      const functionList = this.pkcs11.C_GetFunctionList();
      this.lib = functionList;

      this.logAudit('LOAD_LIBRARY', 'PKCS#11 library loaded successfully', 'INFO');

    } catch (error: any) {
      this.logAudit('LOAD_LIBRARY', `Failed to load library: ${error.message}`, 'CRITICAL');
      
      // Try SoftHSM fallback if enabled
      if (this.config.softHSMFallback && !this.config.libraryPath.includes('softhsm')) {
        this.trySoftHSMFallback();
      } else {
        throw new Error(`Failed to load PKCS#11 library: ${error.message}`);
      }
    }
  }

  /**
   * Attempts to use SoftHSM as a fallback when physical HSM is unavailable
   */
  private async trySoftHSMFallback(): Promise<void> {
    this.logAudit('SOFTHSM_FALLBACK', 'Attempting SoftHSM fallback', 'WARNING');

    const softHSMPaths = [
      '/usr/lib/softhsm/libsofthsm2.so',
      '/usr/lib64/softhsm/libsofthsm2.so',
      '/usr/local/lib/softhsm/libsofthsm2.so',
      '/opt/softhsm/lib/libsofthsm2.so'
    ];

    for (const path of softHSMPaths) {
      try {
        this.config.libraryPath = path;
        this.pkcs11 = new pkcs11.PKCS11();
        this.pkcs11.load(path);
        
        const functionList = this.pkcs11.C_GetFunctionList();
        this.lib = functionList;
        
        this.logAudit('SOFTHSM_FALLBACK', `SoftHSM loaded from ${path}`, 'INFO');
        return;
      } catch {
        continue;
      }
    }

    throw new Error('No suitable PKCS#11 library found (Physical HSM or SoftHSM)');
  }

  /**
   * Opens a session to the HSM slot
   */
  private async openSession(): Promise<void> {
    const startTime = Date.now();

    try {
      // Get slot info first to verify slot exists
      const slots = this.getSlots();
      const slot = slots.find((s: any) =>.slotId === this.config.slotId);

      if (!slot) {
        throw new Error(`Slot ${this.config.slotId} not found. Available slots: ${slots.map((s: any) => s.slotId).join(', ')}`);
      }

      // Open session
      this.session = this.lib.C_OpenSession(
        this.config.slotId,
        pkcs11.CKF_RW_SESSION | (this.config.readOnly ? 0 : pkcs11.CKF_RW_SESSION)
      );

      // Get slot information
      this.slotInfo = this.lib.C_GetSlotInfo(this.config.slotId);

      const duration = Date.now() - startTime;
      this.logAudit('OPEN_SESSION', `Session opened on slot ${this.config.slotId}`, 'INFO', duration);

    } catch (error: any) {
      this.logAudit('OPEN_SESSION', `Failed to open session: ${error.message}`, 'ERROR');
      throw error;
    }
  }

  /**
   * Logs into the HSM with the configured PIN
   */
  private async login(): Promise<void> {
    const startTime = Date.now();

    try {
      // Login as User (CKU_USER)
      this.lib.C_Login(this.session, pkcs11.CKU_USER, this.config.pin);

      const duration = Date.now() - startTime;
      this.logAudit('LOGIN', 'Successfully logged into HSM', 'INFO', duration);

    } catch (error: any) {
      // Handle different login errors
      if (error.errorCode === pkcs11.CKR_PIN_INCORRECT) {
        this.logAudit('LOGIN', 'Incorrect PIN', 'CRITICAL');
        throw new Error('Invalid HSM PIN');
      } else if (error.errorCode === pkcs11.CKR_USER_ALREADY_LOGGED_IN) {
        this.logAudit('LOGIN', 'User already logged in', 'WARNING');
      } else {
        this.logAudit('LOGIN', `Login failed: ${error.message}`, 'CRITICAL');
        throw error;
      }
    }
  }

  /**
   * Lists all available slots
   */
  getSlots(): any[] {
    try {
      const slots = this.lib.C_GetSlotList(true);
      return slots.map((slotId: number, index: number) => ({
        slotId,
        index,
        info: this.lib.C_GetSlotInfo(slotId)
      }));
    } catch (error: any) {
      this.logAudit('GET_SLOTS', `Failed to get slots: ${error.message}`, 'ERROR');
      return [];
    }
  }

  /**
   * Generates a new cryptographic key in the HSM
   * 
   * @param request - Key generation parameters
   * @returns Generated key information
   */
  async generateKey(request: KeyGenerationRequest): Promise<HSMKey> {
    const startTime = Date.now();

    try {
      this.logAudit('GENERATE_KEY', `Generating ${request.type} key: ${request.label}`, 'INFO');

      let keyTemplate: any[];

      // Build key template based on type
      switch (request.type) {
        case KeyType.RSA:
          keyTemplate = this.buildRSATemplate(request);
          break;
        case KeyType.ECDSA:
          keyTemplate = this.buildECDSATemplate(request);
          break;
        case KeyType.AES:
          keyTemplate = this.buildAESTemplate(request);
          break;
        case KeyType.HMAC:
          keyTemplate = this.buildHMACTemplate(request);
          break;
        default:
          throw new Error(`Unsupported key type: ${request.type}`);
      }

      // Generate the key
      const keyHandle = this.lib.C_GenerateKey(
        this.session,
        { mechanism: this.getMechanism(request.algorithm, request.type) },
        keyTemplate
      );

      // Get key attributes
      const keyInfo = this.getKeyInfo(keyHandle, request.label);
      const duration = Date.now() - startTime;

      // Cache the key
      this.keyCache.set(keyInfo.id.toString('hex'), keyInfo);

      this.logAudit('GENERATE_KEY', `Key generated successfully: ${request.label}`, 'INFO', duration);

      return keyInfo;

    } catch (error: any) {
      this.logAudit('GENERATE_KEY', `Key generation failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Builds RSA key generation template
   */
  private buildRSATemplate(request: KeyGenerationRequest): any[] {
    return [
      { type: pkcs11.CKA_LABEL, value: request.label },
      { type: pkcs11.CKA_ID, value: Buffer.from(uuidv4().replace(/-/g, '')) },
      { type: pkcs11.CKA_TOKEN, value: true },
      { type: pkcs11.CKA_PRIVATE, value: !request.extractable },
      { type: pkcs11.CKA_SENSITIVE, value: !request.extractable },
      { type: pkcs11.CKA_EXTRACTABLE, value: request.extractable || false },
      { type: pkcs11.CKA_MODULUS_BITS, value: request.size },
      { type: pkcs11.CKA_PUBLIC_EXPONENT, value: Buffer.from([0x01, 0x00, 0x01]) },
      { type: pkcs11.CKA_ENCRYPT, value: request.usage?.includes(KeyUsage.ENCRYPT) || true },
      { type: pkcs11.CKA_DECRYPT, value: request.usage?.includes(KeyUsage.DECRYPT) || true },
      { type: pkcs11.CKA_SIGN, value: request.usage?.includes(KeyUsage.SIGN) || true },
      { type: pkcs11.CKA_VERIFY, value: request.usage?.includes(KeyUsage.VERIFY) || true },
      { type: pkcs11.CKA_WRAP, value: request.usage?.includes(KeyUsage.WRAP) || false },
      { type: pkcs11.CKA_UNWRAP, value: request.usage?.includes(KeyUsage.UNWRAP) || false }
    ];
  }

  /**
   * Builds ECDSA key generation template
   */
  private buildECDSATemplate(request: KeyGenerationRequest): any[] {
    const curveOids: Record<number, Buffer> = {
      256: Buffer.from([0x06, 0x08, 0x2A, 0x86, 0x48, 0xCE, 0x3D, 0x03, 0x01, 0x07]), // secp256r1
      384: Buffer.from([0x06, 0x05, 0x2B, 0x81, 0x04, 0x00, 0x22]), // secp384r1
      521: Buffer.from([0x06, 0x05, 0x2B, 0x81, 0x04, 0x00, 0x23])  // secp521r1
    };

    return [
      { type: pkcs11.CKA_LABEL, value: request.label },
      { type: pkcs11.CKA_ID, value: Buffer.from(uuidv4().replace(/-/g, '')) },
      { type: pkcs11.CKA_TOKEN, value: true },
      { type: pkcs11.CKA_PRIVATE, value: !request.extractable },
      { type: pkcs11.CKA_SENSITIVE, value: !request.extractable },
      { type: pkcs11.CKA_EXTRACTABLE, value: request.extractable || false },
      { type: pkcs11.CKA_EC_PARAMS, value: curveOids[request.size] || curveOids[256] },
      { type: pkcs11.CKA_SIGN, value: request.usage?.includes(KeyUsage.SIGN) || true },
      { type: pkcs11.CKA_VERIFY, value: request.usage?.includes(KeyUsage.VERIFY) || true }
    ];
  }

  /**
   * Builds AES key generation template
   */
  private buildAESTemplate(request: KeyGenerationRequest): any[] {
    return [
      { type: pkcs11.CKA_LABEL, value: request.label },
      { type: pkcs11.CKA_ID, value: Buffer.from(uuidv4().replace(/-/g, '')) },
      { type: pkcs11.CKA_TOKEN, value: true },
      { type: pkcs11.CKA_PRIVATE, value: true },
      { type: pkcs11.CKA_SENSITIVE, value: true },
      { type: pkcs11.CKA_EXTRACTABLE, value: false },
      { type: pkcs11.CKA_ENCRYPT, value: request.usage?.includes(KeyUsage.ENCRYPT) || true },
      { type: pkcs11.CKA_DECRYPT, value: request.usage?.includes(KeyUsage.DECRYPT) || true },
      { type: pkcs11.CKA_WRAP, value: request.usage?.includes(KeyUsage.WRAP) || true },
      { type: pkcs11.CKA_UNWRAP, value: request.usage?.includes(KeyUsage.UNWRAP) || true },
      { type: pkcs11.CKA_VALUE_LEN, value: request.size / 8 }
    ];
  }

  /**
   * Builds HMAC key generation template
   */
  private buildHMACTemplate(request: KeyGenerationRequest): any[] {
    return [
      { type: pkcs11.CKA_LABEL, value: request.label },
      { type: pkcs11.CKA_ID, value: Buffer.from(uuidv4().replace(/-/g, '')) },
      { type: pkcs11.CKA_TOKEN, value: true },
      { type: pkcs11.CKA_PRIVATE, value: true },
      { type: pkcs11.CKA_SENSITIVE, value: true },
      { type: pkcs11.CKA_EXTRACTABLE, value: false },
      { type: pkcs11.CKA_SIGN, value: request.usage?.includes(KeyUsage.SIGN) || true },
      { type: pkcs11.CKA_VERIFY, value: request.usage?.includes(KeyUsage.VERIFY) || true },
      { type: pkcs11.CKA_VALUE_LEN, value: request.size / 8 }
    ];
  }

  /**
   * Gets the appropriate mechanism for key generation
   */
  private getMechanism(algorithm: string, keyType: KeyType): any {
    const mechanismMap: Record<string, any> = {
      'RSA-SHA256': pkcs11.CKM_RSA_PKCS,
      'RSA-SHA384': pkcs11.CKM_RSA_PKCS,
      'RSA-SHA512': pkcs11.CKM_RSA_PKCS,
      'ECDSA-SHA256': pkcs11.CKM_ECDSA,
      'ECDSA-SHA384': pkcs11.CKM_ECDSA,
      'ECDSA-SHA512': pkcs11.CKM_ECDSA,
      'AES-256-GCM': pkcs11.CKM_AES_GCM,
      'AES-256-CBC': pkcs11.CKM_AES_CBC,
      'HMAC-SHA256': pkcs11.CKM_SHA256_HMAC
    };

    const mechanism = mechanismMap[algorithm] || mechanismMap[`${keyType}-SHA256`];
    if (!mechanism) {
      throw new Error(`Unsupported algorithm: ${algorithm}`);
    }
    return mechanism;
  }

  /**
   * Gets key information from the HSM
   */
  private getKeyInfo(keyHandle: any, label: string): HSMKey {
    // Get key type
    const keyType = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_CLASS);
    
    // Get key ID
    const keyId = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_ID);

    // Get key label
    const keyLabel = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_LABEL) || Buffer.from(label);

    // Get key size
    let keySize = 0;
    if (keyType === pkcs11.CKO_PUBLIC_KEY || keyType === pkcs11.CKO_PRIVATE_KEY) {
      keySize = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_MODULUS_BITS) || 256;
    } else if (keyType === pkcs11.CKO_SECRET_KEY) {
      keySize = (this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_VALUE_LEN) || 32) * 8;
    }

    // Get extractable flag
    const extractable = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_EXTRACTABLE);
    const sensitive = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_SENSITIVE);

    // Determine algorithm and type
    let algorithm = 'Unknown';
    let type = KeyType.RSA;
    
    if (keyType === pkcs11.CKO_PUBLIC_KEY || keyType === pkcs11.CKO_PRIVATE_KEY) {
      if (keyType === pkcs11.CKO_PUBLIC_KEY) {
        const ecParams = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_EC_PARAMS);
        if (ecParams) {
          type = KeyType.ECDSA;
          algorithm = 'ECDSA';
        } else {
          algorithm = 'RSA';
        }
      } else {
        const ecParams = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_EC_PARAMS);
        if (ecParams) {
          type = KeyType.ECDSA;
          algorithm = 'ECDSA';
        } else {
          algorithm = 'RSA';
        }
      }
    } else if (keyType === pkcs11.CKO_SECRET_KEY) {
      const valueLen = this.lib.C_GetObjectAttribute(this.session, keyHandle, pkcs11.CKA_VALUE_LEN) || 32;
      if (valueLen === 32) {
        type = KeyType.HMAC;
        algorithm = 'HMAC-SHA256';
      } else {
        type = KeyType.AES;
        algorithm = 'AES';
      }
    }

    return {
      id: keyId,
      label: keyLabel.toString(),
      type,
      algorithm,
      size: keySize,
      extractable,
      sensitive,
      createdAt: new Date(),
      status: KeyStatus.ACTIVE,
      usage: [KeyUsage.SIGN, KeyUsage.VERIFY]
    };
  }

  /**
   * Lists all keys in the HSM
   */
  async listKeys(): Promise<HSMKey[]> {
    if (!this.isInitialized) {
      throw new Error('HSM not initialized');
    }

    try {
      // Search for all keys
      const template = [{ type: pkcs11.CKA_TOKEN, value: true }];
      const handles = this.lib.C_FindObjectsInit(this.session, template);
      const keyHandles: any[] = [];

      let handle = this.lib.C_FindObjects(this.session, handles, 100);
      while (handle) {
        keyHandles.push(handle);
        handle = this.lib.C_FindObjects(this.session, handles, 100);
      }

      this.lib.C_FindObjectsFinal(this.session, handles);

      // Convert handles to key info
      const keys: HSMKey[] = [];
      for (const keyHandle of keyHandles) {
        try {
          const keyInfo = this.getKeyInfo(keyHandle, '');
          keys.push(keyInfo);
        } catch {
          // Skip objects that can't be read
        }
      }

      return keys;

    } catch (error: any) {
      this.logAudit('LIST_KEYS', `Failed to list keys: ${error.message}`, 'ERROR');
      throw error;
    }
  }

  /**
   * Performs a digital signing operation
   * 
   * @param request - Signing request parameters
   * @returns Signing result with signature
   */
  async sign(request: SigningRequest): Promise<SigningResult> {
    const startTime = Date.now();

    try {
      this.logAudit('SIGN', `Signing data with key: ${request.keyId}`, 'INFO');

      // Find the key
      const key = await this.findKeyById(request.keyId);
      if (!key) {
        throw new Error(`Key not found: ${request.keyId}`);
      }

      // Get key handle
      const keyHandle = await this.getKeyHandle(request.keyId);

      // Initialize signing
      const mechanism = this.getSigningMechanism(request.algorithm);
      this.lib.C_SignInit(this.session, { mechanism }, keyHandle);

      // Perform signing
      let signature: Buffer;
      
      if (request.mode === SigningMode.HASH) {
        // Data is already a hash
        signature = Buffer.from(this.lib.C_Sign(this.session, request.data));
      } else {
        // Hash the data first, then sign
        const hash = crypto.createHash(this.mapHashAlgorithm(request.algorithm)).update(request.data).digest();
        signature = Buffer.from(this.lib.C_Sign(this.session, hash));
      }

      const duration = Date.now() - startTime;
      this.logAudit('SIGN', `Signature generated successfully`, 'INFO', duration);

      return {
        signature,
        keyId: request.keyId,
        timestamp: new Date(),
        algorithm: request.algorithm
      };

    } catch (error: any) {
      this.logAudit('SIGN', `Signing failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Verifies a digital signature
   * 
   * @param keyId - Key ID used for signing
   * @param data - Original data
   * @param signature - Signature to verify
   * @param algorithm - Signing algorithm
   * @returns Verification result
   */
  async verify(keyId: string, data: Buffer, signature: Buffer, algorithm: string): Promise<boolean> {
    const startTime = Date.now();

    try {
      this.logAudit('VERIFY', `Verifying signature with key: ${keyId}`, 'INFO');

      // Find the key
      const keyHandle = await this.getKeyHandle(keyId);

      // Get the public key for verification
      const publicKeyHandle = await this.findPublicKey(keyId);

      // Initialize verification
      const mechanism = this.getSigningMechanism(algorithm);
      this.lib.C_VerifyInit(this.session, { mechanism }, publicKeyHandle);

      // Hash the data
      const hash = crypto.createHash(this.mapHashAlgorithm(algorithm)).update(data).digest();

      // Verify
      const result = this.lib.C_Verify(this.session, hash, signature);

      const duration = Date.now() - startTime;
      this.logAudit('VERIFY', `Verification ${result ? 'successful' : 'failed'}`, result ? 'INFO' : 'WARNING', duration);

      return result === true;

    } catch (error: any) {
      this.logAudit('VERIFY', `Verification failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Gets the signing mechanism for an algorithm
   */
  private getSigningMechanism(algorithm: string): any {
    const mechanismMap: Record<string, any> = {
      'RSA-SHA256': pkcs11.CKM_SHA256_RSA_PKCS,
      'RSA-SHA384': pkcs11.CKM_SHA384_RSA_PKCS,
      'RSA-SHA512': pkcs11.CKM_SHA512_RSA_PKCS,
      'ECDSA-SHA256': pkcs11.CKM_ECDSA,
      'ECDSA-SHA384': pkcs11.CKM_ECDSA,
      'ECDSA-SHA512': pkcs11.CKM_ECDSA,
      'RSASSA-PSS': pkcs11.CKM_RSA_PKCS_PSS,
      'HMAC-SHA256': pkcs11.CKM_SHA256_HMAC
    };

    const mechanism = mechanismMap[algorithm];
    if (!mechanism) {
      throw new Error(`Unsupported signing algorithm: ${algorithm}`);
    }
    return mechanism;
  }

  /**
   * Maps algorithm name to hash algorithm
   */
  private mapHashAlgorithm(algorithm: string): string {
    const map: Record<string, string> = {
      'RSA-SHA256': 'sha256',
      'RSA-SHA384': 'sha384',
      'RSA-SHA512': 'sha512',
      'ECDSA-SHA256': 'sha256',
      'ECDSA-SHA384': 'sha384',
      'ECDSA-SHA512': 'sha512',
      'HMAC-SHA256': 'sha256'
    };

    return map[algorithm] || 'sha256';
  }

  /**
   * Unwraps an encrypted key using a wrapping key from the HSM
   * 
   * @param request - Key unwrapping parameters
   * @returns Unwrapped key information
   */
  async unwrapKey(request: KeyUnwrapRequest): Promise<HSMKey> {
    const startTime = Date.now();

    try {
      this.logAudit('UNWRAP_KEY', `Unwrapping key with wrapping key: ${request.wrappingKeyId}`, 'INFO');

      // Find the wrapping key
      const wrappingKeyHandle = await this.getKeyHandle(request.wrappingKeyId);

      // Build unwrap template
      const unwrapTemplate = [
        { type: pkcs11.CKA_LABEL, value: `Unwrapped-${Date.now()}` },
        { type: pkcs11.CKA_TOKEN, value: true },
        { type: pkcs11.CKA_PRIVATE, value: true },
        { type: pkcs11.CKA_SENSITIVE, value: true },
        { type: pkcs11.CKA_EXTRACTABLE, value: false }
      ];

      // Unwrap the key
      const mechanism = this.getUnwrapMechanism(request.algorithm);
      const unwrappedKeyHandle = this.lib.C_UnwrapKey(
        this.session,
        { mechanism },
        wrappingKeyHandle,
        request.wrappedKey,
        unwrapTemplate
      );

      // Get key info
      const keyInfo = this.getKeyInfo(unwrappedKeyHandle, `Unwrapped-${Date.now()}`);
      const duration = Date.now() - startTime;

      // Cache the key
      this.keyCache.set(keyInfo.id.toString('hex'), keyInfo);

      this.logAudit('UNWRAP_KEY', `Key unwrapped successfully`, 'INFO', duration);

      return keyInfo;

    } catch (error: any) {
      this.logAudit('UNWRAP_KEY', `Key unwrapping failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Gets the unwrap mechanism for an algorithm
   */
  private getUnwrapMechanism(algorithm: string): any {
    const mechanismMap: Record<string, any> = {
      'AES-256-GCM': pkcs11.CKM_AES_GCM,
      'AES-256-CBC': pkcs11.CKM_AES_CBC,
      'RSA-OAEP': pkcs11.CKM_RSA_PKCS_OAEP
    };

    return mechanismMap[algorithm] || pkcs11.CKM_AES_CBC;
  }

  /**
   * Rotates a key by generating a new key and updating references
   * 
   * @param oldKeyId - ID of the key to rotate
   * @returns New key information
   */
  async rotateKey(oldKeyId: string): Promise<HSMKey> {
    const startTime = Date.now();

    try {
      this.logAudit('ROTATE_KEY', `Rotating key: ${oldKeyId}`, 'INFO');

      // Find the old key
      const oldKey = await this.findKeyById(oldKeyId);
      if (!oldKey) {
        throw new Error(`Key not found: ${oldKeyId}`);
      }

      // Generate a new key with similar parameters
      const newKey = await this.generateKey({
        label: `${oldKey.label}-rotated-${Date.now()}`,
        type: oldKey.type,
        algorithm: oldKey.algorithm,
        size: oldKey.size,
        extractable: false,
        usage: oldKey.usage
      });

      // Mark old key as rotating
      oldKey.status = KeyStatus.ROTATING;
      this.keyCache.set(oldKeyId, oldKey);

      // Update key metadata
      await this.updateKeyMetadata(oldKeyId, { status: KeyStatus.ROTATING });

      const duration = Date.now() - startTime;
      this.logAudit('ROTATE_KEY', `Key rotated successfully: ${oldKeyId} -> ${newKey.label}`, 'INFO', duration);

      return newKey;

    } catch (error: any) {
      this.logAudit('ROTATE_KEY', `Key rotation failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Retires a key (marks as inactive but retains for verification)
   * 
   * @param keyId - ID of the key to retire
   */
  async retireKey(keyId: string): Promise<void> {
    const startTime = Date.now();

    try {
      this.logAudit('RETIRE_KEY', `Retiring key: ${keyId}`, 'INFO');

      // Update key status
      await this.updateKeyMetadata(keyId, { status: KeyStatus.RETIRED });

      // Update cache
      const key = this.keyCache.get(keyId);
      if (key) {
        key.status = KeyStatus.RETIRED;
        this.keyCache.set(keyId, key);
      }

      const duration = Date.now() - startTime;
      this.logAudit('RETIRE_KEY', `Key retired: ${keyId}`, 'INFO', duration);

    } catch (error: any) {
      this.logAudit('RETIRE_KEY', `Key retirement failed: ${error.message}`, 'ERROR', startTime);
      throw error;
    }
  }

  /**
   * Finds a key by its ID
   */
  private async findKeyById(keyId: string): Promise<HSMKey | undefined> {
    // Check cache first
    let key = this.keyCache.get(keyId);
    if (key) {
      return key;
    }

    // Search in HSM
    const keys = await this.listKeys();
    key = keys.find(k => k.id.toString('hex') === keyId);
    
    if (key) {
      this.keyCache.set(keyId, key);
    }
    
    return key;
  }

  /**
   * Gets the key handle for a key ID
   */
  private async getKeyHandle(keyId: string): Promise<any> {
    const template = [
      { type: pkcs11.CKA_ID, value: Buffer.from(keyId, 'hex') },
      { type: pkcs11.CKA_TOKEN, value: true }
    ];

    const handles = this.lib.C_FindObjectsInit(this.session, template);
    const handle = this.lib.C_FindObjects(this.session, handles, 1);
    this.lib.C_FindObjectsFinal(this.session, handles);

    if (!handle) {
      throw new Error(`Key handle not found for ID: ${keyId}`);
    }

    return handle;
  }

  /**
   * Finds the public key corresponding to a private key
   */
  private async findPublicKey(keyId: string): Promise<any> {
    const template = [
      { type: pkcs11.CKA_ID, value: Buffer.from(keyId, 'hex') },
      { type: pkcs11.CKA_CLASS, value: pkcs11.CKO_PUBLIC_KEY },
      { type: pkcs11.CKA_TOKEN, value: true }
    ];

    const handles = this.lib.C_FindObjectsInit(this.session, template);
    const handle = this.lib.C_FindObjects(this.session, handles, 1);
    this.lib.C_FindObjectsFinal(this.session, handles);

    if (!handle) {
      throw new Error(`Public key not found for key ID: ${keyId}`);
    }

    return handle;
  }

  /**
   * Updates key metadata
   */
  private async updateKeyMetadata(keyId: string, metadata: Partial<HSMKey>): Promise<void> {
    const keyHandle = await this.getKeyHandle(keyId);

    // Update attributes
    if (metadata.status !== undefined) {
      // Status is tracked in metadata, not in HSM
      this.logAudit('UPDATE_METADATA', `Updated metadata for key: ${keyId}`, 'INFO');
    }
  }

  /**
   * Refreshes the key and certificate caches
   */
  private async refreshCache(): Promise<void> {
    try {
      this.logAudit('REFRESH_CACHE', 'Refreshing key and certificate cache', 'INFO');

      // List and cache all keys
      const keys = await this.listKeys();
      for (const key of keys) {
        this.keyCache.set(key.id.toString('hex'), key);
      }

      this.logAudit('REFRESH_CACHE', `Cache refreshed: ${keys.length} keys`, 'INFO');

    } catch (error: any) {
      this.logAudit('REFRESH_CACHE', `Cache refresh failed: ${error.message}`, 'ERROR');
    }
  }

  /**
   * Logs an audit entry
   */
  private logAudit(
    operation: string,
    message: string,
    severity: LogSeverity,
    duration?: number,
    keyId?: string
  ): void {
    const entry: AuditLogEntry = {
      id: uuidv4(),
      timestamp: new Date(),
      operation,
      keyId,
      slotId: this.config.slotId,
      result: message.includes('failed') || message.includes('Failed') ? 'FAILURE' : 'SUCCESS',
      errorMessage: message.includes('failed') ? message : undefined,
      actorId: 'HSM-CLIENT',
      severity,
      duration: duration || 0
    };

    this.auditLogs.push(entry);
    this.emit('audit', entry);

    // Console logging for debugging
    const timestamp = entry.timestamp.toISOString();
    console.log(`[HSM-AUDIT] ${timestamp} [${severity}] ${operation}: ${message}`);
  }

  /**
   * Gets audit logs
   */
  getAuditLogs(): AuditLogEntry[] {
    return [...this.auditLogs];
  }

  /**
   * Clears audit logs
   */
  clearAuditLogs(): void {
    this.auditLogs = [];
  }

  /**
   * Gets the HSM status
   */
  getStatus(): { initialized: boolean; slotId: number; slotInfo: any; keyCount: number } {
    return {
      initialized: this.isInitialized,
      slotId: this.config.slotId,
      slotInfo: this.slotInfo,
      keyCount: this.keyCache.size
    };
  }

  /**
   * Handles initialization errors with automatic retry
   */
  private async handleInitializationError(error: any): Promise<void> {
    if (this.connectionRetryCount < this.maxRetries) {
      this.connectionRetryCount++;
      this.logAudit('RETRY', `Retrying initialization (attempt ${this.connectionRetryCount}/${this.maxRetries})`, 'WARNING');
      
      await new Promise(resolve => setTimeout(resolve, this.reconnectDelay));
      
      try {
        await this.initialize();
        this.connectionRetryCount = 0;
      } catch {
        // Error will be re-thrown
      }
    } else {
      this.emit('initializationFailed', error);
    }
  }

  /**
   * Closes the HSM session and cleans up resources
   */
  async close(): Promise<void> {
    try {
      this.logAudit('CLOSE', 'Closing HSM session', 'INFO');

      if (this.session) {
        this.lib.C_Logout(this.session);
        this.lib.C_CloseSession(this.session);
      }

      this.isInitialized = false;
      this.session = null;
      this.keyCache.clear();
      this.certificateCache.clear();

      this.logAudit('CLOSE', 'HSM session closed', 'INFO');
      this.emit('closed');

    } catch (error: any) {
      this.logAudit('CLOSE', `Error closing HSM: ${error.message}`, 'ERROR');
    }
  }
}

// Export types for external use
export {
  HSMConfig,
  HSMKey,
  Certificate,
  SigningRequest,
  SigningResult,
  KeyGenerationRequest,
  KeyUnwrapRequest,
  KeyType,
  KeyStatus,
  KeyUsage,
  CertificateStatus,
  SigningMode,
  LogSeverity,
  AuditLogEntry
};
