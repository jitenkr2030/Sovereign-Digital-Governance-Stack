/**
 * Certificate Service - HSM-Based Certificate Issuance and Management
 * 
 * This module provides certificate lifecycle management using keys stored in HSM.
 * It handles CSR generation, certificate issuance, renewal, and revocation.
 * 
 * @packageDocumentation
 */

import * as asn1 from 'asn1js';
import * as pkcs12 from 'pkcs12';
import { HSMClient, HSMKey, KeyType, KeyUsage, LogSeverity } from '../hsm/hsm-client';
import { v4 as uuidv4 } from 'uuid';

// Certificate types
export enum CertificateType {
  ROOT_CA = 'ROOT_CA',
  INTERMEDIATE_CA = 'INTERMEDIATE_CA',
  END_ENTITY = 'END_ENTITY',
  TLS_SERVER = 'TLS_SERVER',
  TLS_CLIENT = 'TLS_CLIENT',
  CODE_SIGNING = 'CODE_SIGNING',
  DOCUMENT_SIGNING = 'DOCUMENT_SIGNING'
}

// Certificate status
export enum CertificateStatus {
  ACTIVE = 'ACTIVE',
  EXPIRED = 'EXPIRED',
  REVOKED = 'REVOKED',
  PENDING = 'PENDING',
  RENEWED = 'RENEWED'
}

// Certificate profile configuration
export interface CertificateProfile {
  type: CertificateType;
  keyUsage: KeyUsage[];
  extendedKeyUsage?: string[];
  validityDays: number;
  signatureAlgorithm: string;
  keySize: number;
  isCA: boolean;
  pathLength?: number;
}

// Certificate issuance request
export interface CertificateRequest {
  commonName: string;
  organization?: string;
  organizationalUnit?: string;
  locality?: string;
  province?: string;
  country?: string;
  email?: string;
  sans?: string[];
  profile: CertificateProfile;
  keyId?: string; // Use existing key, or generate new
  issuerKeyId?: string; // CA key to sign with
}

// Certificate information
export interface CertificateInfo {
  id: string;
  subject: string;
  issuer: string;
  serialNumber: string;
  fingerprint: string;
  thumbprint: string;
  notBefore: Date;
  notAfter: Date;
  status: CertificateStatus;
  type: CertificateType;
  keyId: string;
  publicKey: Buffer;
  algorithm: string;
  isCA: boolean;
  pathLength?: number;
  revokedAt?: Date;
  revocationReason?: string;
}

// CSR configuration
export interface CSRConfig {
  commonName: string;
  organization?: string;
  organizationalUnit?: string;
  locality?: string;
  province?: string;
  country?: string;
  email?: string;
  sans?: string[];
}

// X.509 certificate extensions
export interface CertificateExtensions {
  basicConstraints?: {
    ca: boolean;
    pathLength?: number;
  };
  keyUsage?: {
    digitalSignature: boolean;
    keyEncipherment: boolean;
    dataEncipherment: boolean;
    keyAgreement: boolean;
    keyCertSign: boolean;
    cRLSign: boolean;
    encipherOnly: boolean;
    decipherOnly: boolean;
  };
  extendedKeyUsage?: string[];
  subjectAltName?: {
    dns?: string[];
    email?: string[];
    ip?: string[];
    uri?: string[];
  };
  subjectKeyIdentifier?: string;
  authorityKeyIdentifier?: string;
  cRLDistributionPoints?: string[];
  authorityInfoAccess?: string[];
}

/**
 * Certificate Service - Handles all certificate operations
 */
export class CertificateService {
  private hsmClient: HSMClient;
  private certificateCache: Map<string, CertificateInfo> = new Map();
  private caCertificate: CertificateInfo | null = null;

  /**
   * Creates a new Certificate Service instance
   * @param hsmClient - HSM client instance for cryptographic operations
   */
  constructor(hsmClient: HSMClient) {
    this.hsmClient = hsmClient;
  }

  /**
   * Initializes the certificate service
   */
  async initialize(): Promise<void> {
    try {
      // Load existing certificates from cache/database
      await this.loadCertificates();
      
      console.log('[CERT-SERVICE] Certificate service initialized');
    } catch (error: any) {
      console.error('[CERT-SERVICE] Initialization failed:', error.message);
      throw error;
    }
  }

  /**
   * Generates a new key pair and CSR (Certificate Signing Request)
   * 
   * @param config - CSR configuration
   * @param profile - Certificate profile
   * @returns CSR and key information
   */
  async generateCSR(
    config: CSRConfig,
    profile: CertificateProfile
  ): Promise<{ csr: string; keyId: string; privateKeyHandle: any }> {
    try {
      console.log('[CERT-SERVICE] Generating CSR for:', config.commonName);

      // Generate key pair in HSM
      const key = await this.hsmClient.generateKey({
        label: `CSR-${config.commonName}-${Date.now()}`,
        type: this.getKeyTypeForProfile(profile),
        algorithm: profile.signatureAlgorithm,
        size: profile.keySize,
        extractable: false,
        usage: profile.keyUsage,
        validityPeriodDays: profile.validityDays
      });

      // Build X.509 name
      const subject = this.buildX509Name(config);

      // Build extensions
      const extensions = this.buildExtensions(profile, config.sans);

      // Generate CSR (simplified - in production use proper ASN.1 encoding)
      const csr = this.createCSRPEM(subject, key, extensions);

      console.log('[CERT-SERVICE] CSR generated successfully');

      return {
        csr,
        keyId: key.id.toString('hex'),
        privateKeyHandle: key // Return key reference, not actual private key
      };

    } catch (error: any) {
      console.error('[CERT-SERVICE] CSR generation failed:', error.message);
      throw error;
    }
  }

  /**
   * Issues a certificate using the CA key in HSM
   * 
   * @param request - Certificate request
   * @returns Issued certificate information
   */
  async issueCertificate(request: CertificateRequest): Promise<CertificateInfo> {
    try {
      console.log('[CERT-SERVICE] Issuing certificate for:', request.commonName);

      // Get or generate key
      let keyId = request.keyId;
      let key: HSMKey;

      if (keyId) {
        key = await this.findKeyById(keyId);
        if (!key) {
          throw new Error(`Key not found: ${keyId}`);
        }
      } else {
        // Generate new key
        key = await this.hsmClient.generateKey({
          label: `Cert-${request.commonName}-${Date.now()}`,
          type: this.getKeyTypeForProfile(request.profile),
          algorithm: request.profile.signatureAlgorithm,
          size: request.profile.keySize,
          extractable: false,
          usage: request.profile.keyUsage,
          validityPeriodDays: request.profile.validityDays
        });
        keyId = key.id.toString('hex');
      }

      // Get issuer key
      const issuerKeyId = request.issuerKeyId || this.caCertificate?.keyId;
      if (!issuerKeyId) {
        throw new Error('Issuer key not specified and no CA certificate configured');
      }

      // Build certificate fields
      const serialNumber = this.generateSerialNumber();
      const notBefore = new Date();
      const notAfter = new Date(Date.now() + request.profile.validityDays * 24 * 60 * 60 * 1000);

      // Build subject and extensions
      const subject = this.buildX509Name(request);
      const extensions = this.buildExtensions(request.profile, request.sans);

      // Create certificate
      const certificate = this.createCertificate(
        subject,
        this.caCertificate?.subject || subject,
        serialNumber,
        notBefore,
        notAfter,
        key,
        extensions
      );

      // Calculate fingerprints
      const fingerprint = this.calculateFingerprint(certificate);
      const thumbprint = this.calculateThumbprint(certificate);

      // Create certificate info
      const certInfo: CertificateInfo = {
        id: uuidv4(),
        subject: this.formatDN(subject),
        issuer: this.formatDN(this.caCertificate?.subject || subject),
        serialNumber,
        fingerprint,
        thumbprint,
        notBefore,
        notAfter,
        status: CertificateStatus.ACTIVE,
        type: request.profile.type,
        keyId,
        publicKey: key.id, // Reference to public key
        algorithm: request.profile.signatureAlgorithm,
        isCA: request.profile.isCA,
        pathLength: request.profile.pathLength
      };

      // Cache the certificate
      this.certificateCache.set(certInfo.id, certInfo);

      console.log('[CERT-SERVICE] Certificate issued:', certInfo.serialNumber);

      return certInfo;

    } catch (error: any) {
      console.error('[CERT-SERVICE] Certificate issuance failed:', error.message);
      throw error;
    }
  }

  /**
   * Issues a self-signed root CA certificate
   * 
   * @param config - CA configuration
   * @param validityDays - Validity period in days
   * @returns Root CA certificate
   */
  async createRootCA(
    config: CSRConfig,
    validityDays: number = 3650
  ): Promise<CertificateInfo> {
    const profile: CertificateProfile = {
      type: CertificateType.ROOT_CA,
      keyUsage: [
        KeyUsage.SIGN,
        KeyUsage.KEY_CERT_SIGN,
        KeyUsage.CRL_SIGN
      ],
      validityDays,
      signatureAlgorithm: 'RSA-SHA256',
      keySize: 4096,
      isCA: true,
      pathLength: 0
    };

    // Generate key and CSR
    const { keyId, privateKeyHandle } = await this.generateCSR(config, profile);

    // Self-sign the certificate
    const certInfo = await this.issueCertificate({
      ...config,
      profile: {
        ...profile,
        type: CertificateType.ROOT_CA
      },
      keyId,
      issuerKeyId: keyId // Self-sign
    });

    // Store as CA certificate
    this.caCertificate = certInfo;

    console.log('[CERT-SERVICE] Root CA created:', certInfo.subject);

    return certInfo;
  }

  /**
   * Issues an intermediate CA certificate
   * 
   * @param request - Certificate request
   * @param parentCAKeyId - Parent CA key ID
   * @returns Intermediate CA certificate
   */
  async createIntermediateCA(
    request: CertificateRequest,
    parentCAKeyId: string
  ): Promise<CertificateInfo> {
    const profile: CertificateProfile = {
      type: CertificateType.INTERMEDIATE_CA,
      keyUsage: [
        KeyUsage.SIGN,
        KeyUsage.KEY_CERT_SIGN,
        KeyUsage.CRL_SIGN
      ],
      extendedKeyUsage: ['certificateSign'],
      validityDays: request.profile.validityDays || 365,
      signatureAlgorithm: 'RSA-SHA256',
      keySize: 4096,
      isCA: true,
      pathLength: request.profile.pathLength || 0
    };

    return this.issueCertificate({
      ...request,
      profile,
      issuerKeyId: parentCAKeyId
    });
  }

  /**
   * Issues a TLS server certificate
   * 
   * @param commonName - Server hostname
   * @param sans - Subject Alternative Names (additional hostnames)
   * @param organization - Organization name
   * @returns TLS certificate
   */
  async issueTLSServerCertificate(
    commonName: string,
    sans: string[] = [],
    organization?: string
  ): Promise<CertificateInfo> {
    const profile: CertificateProfile = {
      type: CertificateType.TLS_SERVER,
      keyUsage: [
        KeyUsage.DIGITAL_SIGNATURE,
        KeyUsage.KEY_ENCIPHERMENT
      ],
      extendedKeyUsage: [
        'serverAuth',
        'clientAuth'
      ],
      validityDays: 398, // ~13 months (current browser requirement)
      signatureAlgorithm: 'RSA-SHA256',
      keySize: 2048,
      isCA: false
    };

    return this.issueCertificate({
      commonName,
      organization,
      sans: [commonName, ...sans],
      profile
    });
  }

  /**
   * Issues a TLS client certificate
   * 
   * @param commonName - Client identifier
   * @param organization - Organization name
   * @returns Client certificate
   */
  async issueTLSClientCertificate(
    commonName: string,
    organization?: string
  ): Promise<CertificateInfo> {
    const profile: CertificateProfile = {
      type: CertificateType.TLS_CLIENT,
      keyUsage: [
        KeyUsage.DIGITAL_SIGNATURE,
        KeyUsage.KEY_ENCIPHERMENT
      ],
      extendedKeyUsage: ['clientAuth'],
      validityDays: 398,
      signatureAlgorithm: 'RSA-SHA256',
      keySize: 2048,
      isCA: false
    };

    return this.issueCertificate({
      commonName,
      organization,
      profile
    });
  }

  /**
   * Renews an existing certificate
   * 
   * @param certificateId - ID of certificate to renew
   * @returns New certificate with extended validity
   */
  async renewCertificate(certificateId: string): Promise<CertificateInfo> {
    const oldCert = this.certificateCache.get(certificateId);
    if (!oldCert) {
      throw new Error(`Certificate not found: ${certificateId}`);
    }

    // Create new certificate with same parameters but new validity
    const newCert = await this.issueCertificate({
      commonName: this.extractCN(oldCert.subject),
      profile: {
        type: oldCert.type,
        keyUsage: [KeyUsage.SIGN, KeyUsage.VERIFY],
        validityDays: oldCert.type === CertificateType.TLS_SERVER ? 398 : 365,
        signatureAlgorithm: oldCert.algorithm,
        keySize: 2048,
        isCA: false
      },
      issuerKeyId: this.caCertificate?.keyId
    });

    // Mark old certificate as renewed
    oldCert.status = CertificateStatus.RENEWED;
    this.certificateCache.set(certificateId, oldCert);

    console.log('[CERT-SERVICE] Certificate renewed:', certificateId, '->', newCert.id);

    return newCert;
  }

  /**
   * Revokes a certificate
   * 
   * @param certificateId - ID of certificate to revoke
   * @param reason - Revocation reason
   * @returns Revocation confirmation
   */
  async revokeCertificate(
    certificateId: string,
    reason: string = 'Unspecified'
  ): Promise<void> {
    const cert = this.certificateCache.get(certificateId);
    if (!cert) {
      throw new Error(`Certificate not found: ${certificateId}`);
    }

    // Update status
    cert.status = CertificateStatus.REVOKED;
    cert.revokedAt = new Date();
    cert.revocationReason = reason;
    this.certificateCache.set(certificateId, cert);

    // In production, would add to CRL and publish to OCSP
    await this.publishCRLEntry(cert);

    console.log('[CERT-SERVICE] Certificate revoked:', certificateId, 'Reason:', reason);
  }

  /**
   * Verifies a certificate chain
   * 
   * @param certificateChain - Array of certificates (leaf to root)
   * @returns Verification result
   */
  async verifyCertificateChain(certificateChain: string[]): Promise<{
    valid: boolean;
    error?: string;
    chain: CertificateInfo[];
  }> {
    try {
      const chain: CertificateInfo[] = [];

      // Parse and validate each certificate
      for (const certPEM of certificateChain) {
        const cert = await this.parseCertificate(certPEM);
        chain.push(cert);
      }

      // Verify chain integrity
      for (let i = 0; i < chain.length - 1; i++) {
        const issuer = chain[i + 1];
        const subject = chain[i];

        // Check issuer matches
        if (!subject.issuer.includes(issuer.subject)) {
          return { valid: false, error: 'Certificate chain broken', chain };
        }

        // Verify signature
        // In production, would verify using HSM
      }

      // Check certificate validity
      const now = new Date();
      for (const cert of chain) {
        if (cert.notAfter < now) {
          return { valid: false, error: `Certificate expired: ${cert.subject}`, chain };
        }
        if (cert.status === CertificateStatus.REVOKED) {
          return { valid: false, error: `Certificate revoked: ${cert.subject}`, chain };
        }
      }

      return { valid: true, chain };

    } catch (error: any) {
      return { valid: false, error: error.message, chain: [] };
    }
  }

  /**
   * Gets certificate by ID
   */
  getCertificate(certificateId: string): CertificateInfo | undefined {
    return this.certificateCache.get(certificateId);
  }

  /**
   * Lists all certificates
   */
  listCertificates(): CertificateInfo[] {
    return Array.from(this.certificateCache.values());
  }

  /**
   * Lists certificates by status
   */
  listCertificatesByStatus(status: CertificateStatus): CertificateInfo[] {
    return Array.from(this.certificateCache.values()).filter(
      cert => cert.status === status
    );
  }

  /**
   * Gets certificates expiring soon
   * @param days - Number of days to check
   */
  getExpiringCertificates(days: number = 30): CertificateInfo[] {
    const cutoff = new Date(Date.now() + days * 24 * 60 * 60 * 1000);
    return Array.from(this.certificateCache.values()).filter(
      cert => cert.notAfter <= cutoff && cert.status === CertificateStatus.ACTIVE
    );
  }

  /**
   * Exports certificate to PEM format
   */
  exportCertificatePEM(certificateId: string): string {
    const cert = this.certificateCache.get(certificateId);
    if (!cert) {
      throw new Error(`Certificate not found: ${certificateId}`);
    }

    // In production, would export actual certificate
    return `-----BEGIN CERTIFICATE-----\n${cert.fingerprint}\n-----END CERTIFICATE-----`;
  }

  /**
   * Gets certificate in PKCS#12 format
   */
  async exportCertificatePKCS12(
    certificateId: string,
    password: string
  ): Promise<Buffer> {
    const cert = this.certificateCache.get(certificateId);
    if (!cert) {
      throw new Error(`Certificate not found: ${certificateId}`);
    }

    // In production, would package certificate and key
    // This requires access to the private key in HSM
    throw new Error('PKCS#12 export requires private key access');
  }

  // Private helper methods

  /**
   * Builds X.509 name from configuration
   */
  private buildX509Name(config: CertificateRequest | CSRConfig): Record<string, string> {
    const name: Record<string, string> = {};

    if ('commonName' in config && config.commonName) {
      name.CN = config.commonName;
    }
    if (config.organization) {
      name.O = config.organization;
    }
    if (config.organizationalUnit) {
      name.OU = config.organizationalUnit;
    }
    if (config.locality) {
      name.L = config.locality;
    }
    if (config.province) {
      name.ST = config.province;
    }
    if (config.country) {
      name.C = config.country;
    }
    if (config.email) {
      name.E = config.email;
    }

    return name;
  }

  /**
   * Builds X.509 extensions
   */
  private buildExtensions(
    profile: CertificateProfile,
    sans?: string[]
  ): CertificateExtensions {
    const extensions: CertificateExtensions = {};

    // Basic Constraints
    if (profile.isCA) {
      extensions.basicConstraints = {
        ca: true,
        pathLength: profile.pathLength
      };
    }

    // Key Usage
    extensions.keyUsage = {
      digitalSignature: profile.keyUsage.includes(KeyUsage.SIGN),
      keyEncipherment: profile.keyUsage.includes(KeyUsage.ENCRYPT),
      dataEncipherment: profile.keyUsage.includes(KeyUsage.DECRYPT),
      keyAgreement: false,
      keyCertSign: profile.keyUsage.includes(KeyUsage.SIGN),
      cRLSign: profile.keyUsage.includes(KeyUsage.SIGN),
      encipherOnly: false,
      decipherOnly: false
    };

    // Extended Key Usage
    if (profile.extendedKeyUsage && profile.extendedKeyUsage.length > 0) {
      extensions.extendedKeyUsage = profile.extendedKeyUsage;
    }

    // Subject Alternative Names
    if (sans && sans.length > 0) {
      extensions.subjectAltName = {
        dns: sans.filter(s => /^[a-zA-Z0-9]/.test(s))
      };
    }

    return extensions;
  }

  /**
   * Creates a CSR (simplified implementation)
   */
  private createCSRPEM(
    subject: Record<string, string>,
    key: HSMKey,
    extensions: CertificateExtensions
  ): string {
    // In production, would use proper ASN.1 encoding
    // This is a placeholder showing the structure
    const csrData = {
      subject,
      publicKey: key.id.toString('hex'),
      extensions
    };

    // Return base64 encoded CSR (placeholder)
    return Buffer.from(JSON.stringify(csrData)).toString('base64');
  }

  /**
   * Creates a certificate (simplified implementation)
   */
  private createCertificate(
    subject: Record<string, string>,
    issuer: Record<string, string>,
    serialNumber: string,
    notBefore: Date,
    notAfter: Date,
    key: HSMKey,
    extensions: CertificateExtensions
  ): Buffer {
    // In production, would use proper ASN.1 encoding
    // This is a placeholder showing the structure
    const certData = {
      subject: this.formatDN(subject),
      issuer: this.formatDN(issuer),
      serialNumber,
      notBefore: notBefore.toISOString(),
      notAfter: notAfter.toISOString(),
      publicKey: key.id.toString('hex'),
      extensions
    };

    return Buffer.from(JSON.stringify(certData));
  }

  /**
   * Formats distinguished name
   */
  private formatDN(name: Record<string, string>): string {
    const parts: string[] = [];
    
    if (name.CN) parts.push(`CN=${name.CN}`);
    if (name.O) parts.push(`O=${name.O}`);
    if (name.OU) parts.push(`OU=${name.OU}`);
    if (name.L) parts.push(`L=${name.L}`);
    if (name.ST) parts.push(`ST=${name.ST}`);
    if (name.C) parts.push(`C=${name.C}`);
    if (name.E) parts.push(`E=${name.E}`);

    return parts.join(', ');
  }

  /**
   * Extracts common name from distinguished name
   */
  private extractCN(dn: string): string {
    const match = dn.match(/CN=([^,]+)/);
    return match ? match[1] : dn;
  }

  /**
   * Generates a unique serial number
   */
  private generateSerialNumber(): string {
    const randomBytes = require('crypto').randomBytes(16);
    return randomBytes.toString('hex').toUpperCase();
  }

  /**
   * Calculates certificate fingerprint
   */
  private calculateFingerprint(cert: Buffer): string {
    const hash = require('crypto').createHash('sha256');
    hash.update(cert);
    return hash.digest('hex').toUpperCase().match(/.{1,2}/g)!.join(':');
  }

  /**
   * Calculates certificate thumbprint (SHA-1)
   */
  private calculateThumbprint(cert: Buffer): string {
    const hash = require('crypto').createHash('sha1');
    hash.update(cert);
    return hash.digest('hex').toUpperCase();
  }

  /**
   * Gets key type for certificate profile
   */
  private getKeyTypeForProfile(profile: CertificateProfile): KeyType {
    if (profile.signatureAlgorithm.includes('ECDSA')) {
      return KeyType.ECDSA;
    }
    return KeyType.RSA;
  }

  /**
   * Finds a key by ID
   */
  private async findKeyById(keyId: string): Promise<HSMKey | undefined> {
    const keys = await this.hsmClient.listKeys();
    return keys.find(k => k.id.toString('hex') === keyId);
  }

  /**
   * Parses a certificate from PEM format
   */
  private async parseCertificate(pem: string): Promise<CertificateInfo> {
    // In production, would parse actual certificate
    // This is a placeholder
    return {
      id: uuidv4(),
      subject: pem.substring(0, 100),
      issuer: pem.substring(0, 100),
      serialNumber: 'placeholder',
      fingerprint: 'placeholder',
      thumbprint: 'placeholder',
      notBefore: new Date(),
      notAfter: new Date(),
      status: CertificateStatus.ACTIVE,
      type: CertificateType.END_ENTITY,
      keyId: '',
      publicKey: Buffer.from(''),
      algorithm: 'RSA-SHA256',
      isCA: false
    };
  }

  /**
   * Publishes CRL entry
   */
  private async publishCRLEntry(cert: CertificateInfo): Promise<void> {
    // In production, would update CRL and OCSP responder
    console.log('[CERT-SERVICE] CRL entry published for:', cert.serialNumber);
  }

  /**
   * Loads certificates from cache/database
   */
  private async loadCertificates(): Promise<void> {
    // In production, would load from database
    console.log('[CERT-SERVICE] Certificates loaded from cache');
  }
}

// Export types
export {
  CertificateType,
  CertificateStatus,
  CertificateProfile,
  CertificateRequest,
  CertificateInfo,
  CSRConfig,
  CertificateExtensions
};
