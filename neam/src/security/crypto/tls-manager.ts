/**
 * TLS Certificate Manager - Handles TLS certificate lifecycle
 * 
 * This module manages TLS certificates including issuance, renewal,
 * deployment, and monitoring for the National Dashboard.
 * 
 * @packageDocumentation
 */

import * as tls from 'tls';
import * as fs from 'fs';
import * as path from 'path';
import { EventEmitter } from 'events';
import { v4 as uuidv4 } from 'uuid';
import {
  CertificateService,
  CertificateType,
  CertificateStatus,
  CertificateInfo,
  CertificateProfile
} from './certificate-service';
import { HSMClient } from '../hsm/hsm-client';

// TLS Configuration
export interface TLSConfig {
  enabled: boolean;
  port: number;
  minVersion: string;
  maxVersion: string;
  ciphers: string[];
  honorCipherOrder: boolean;
  requestClientCert: boolean;
  rejectUnauthorized: boolean;
}

// Certificate deployment configuration
export interface CertificateDeployment {
  certificateId: string;
  privateKeyId: string;
  chainCertificateIds?: string[];
  deploymentPath: string;
  reloadCommand?: string;
  restartService?: string;
}

// Certificate monitoring configuration
export interface MonitoringConfig {
  checkIntervalMinutes: number;
  warningThresholdDays: number;
  criticalThresholdDays: number;
  autoRenewEnabled: boolean;
  notificationEnabled: boolean;
  notificationRecipients?: string[];
}

// Certificate inventory item
export interface CertificateInventory {
  id: string;
  commonName: string;
  type: CertificateType;
  status: CertificateStatus;
  issuer: string;
  serialNumber: string;
  fingerprint: string;
  notBefore: Date;
  notAfter: Date;
  daysUntilExpiry: number;
  deployedServices: string[];
  lastChecked: Date;
  autoRenew: boolean;
}

// Service endpoint configuration
export interface ServiceEndpoint {
  name: string;
  host: string;
  port: number;
  protocol: 'https' | 'wss';
  certificateId?: string;
  enabled: boolean;
}

// TLS monitoring event
export interface TLSMonitoringEvent {
  type: 'expiring' | 'expired' | 'deployed' | 'renewed' | 'error';
  certificateId: string;
  commonName: string;
  daysUntilExpiry?: number;
  message: string;
  timestamp: Date;
}

/**
 * TLS Certificate Manager
 */
export class TLSCertificateManager extends EventEmitter {
  private certificateService: CertificateService;
  private hsmClient: HSMClient;
  private config: TLSConfig;
  private monitoringConfig: MonitoringConfig;
  private inventory: Map<string, CertificateInventory> = new Map();
  private services: Map<string, ServiceEndpoint> = new Map();
  private monitoringInterval: NodeJS.Timeout | null = null;
  private serverInstances: Map<string, tls.Server> = new Map();

  /**
   * Creates a new TLS Certificate Manager
   */
  constructor(
    certificateService: CertificateService,
    hsmClient: HSMClient,
    config?: Partial<TLSConfig>,
    monitoringConfig?: Partial<MonitoringConfig>
  ) {
    super();
    this.certificateService = certificateService;
    this.hsmClient = hsmClient;

    // Default TLS configuration
    this.config = {
      enabled: true,
      port: 443,
      minVersion: 'TLSv1.2',
      maxVersion: 'TLSv1.3',
      ciphers: [
        'TLS_AES_256_GCM_SHA384',
        'TLS_AES_128_GCM_SHA256',
        'TLS_CHACHA20_POLY1305_SHA256',
        'ECDHE-RSA-AES256-GCM-SHA384',
        'ECDHE-RSA-AES128-GCM-SHA256',
        'ECDHE-RSA-AES256-SHA384',
        'ECDHE-RSA-AES128-SHA256'
      ].join(':'),
      honorCipherOrder: true,
      requestClientCert: false,
      rejectUnauthorized: true,
      ...config
    };

    // Default monitoring configuration
    this.monitoringConfig = {
      checkIntervalMinutes: 60,
      warningThresholdDays: 30,
      criticalThresholdDays: 7,
      autoRenewEnabled: false,
      notificationEnabled: true,
      ...monitoringConfig
    };
  }

  /**
   * Initializes the TLS manager
   */
  async initialize(): Promise<void> {
    try {
      console.log('[TLS-MANAGER] Initializing TLS Certificate Manager');

      // Initialize certificate service
      await this.certificateService.initialize();

      // Load certificate inventory
      await this.loadInventory();

      // Start monitoring
      this.startMonitoring();

      console.log('[TLS-MANAGER] TLS Certificate Manager initialized');
    } catch (error: any) {
      console.error('[TLS-MANAGER] Initialization failed:', error.message);
      throw error;
    }
  }

  /**
   * Issues a new TLS server certificate
   * 
   * @param commonName - Server hostname
   * @param sans - Additional hostnames (SANs)
   * @param organization - Organization name
   * @returns Issued certificate information
   */
  async issueTLSCertificate(
    commonName: string,
    sans: string[] = [],
    organization?: string
  ): Promise<CertificateInfo> {
    console.log('[TLS-MANAGER] Issuing TLS certificate for:', commonName);

    const cert = await this.certificateService.issueTLSServerCertificate(
      commonName,
      sans,
      organization
    );

    // Add to inventory
    await this.addToInventory(cert);

    // Emit event
    this.emit('certificateIssued', {
      type: 'deployed',
      certificateId: cert.id,
      commonName: cert.subject,
      message: `TLS certificate issued for ${commonName}`
    });

    return cert;
  }

  /**
   * Deploys a certificate to a service endpoint
   * 
   * @param deployment - Certificate deployment configuration
   */
  async deployCertificate(deployment: CertificateDeployment): Promise<void> {
    try {
      console.log('[TLS-MANAGER] Deploying certificate:', deployment.certificateId);

      const cert = this.certificateService.getCertificate(deployment.certificateId);
      if (!cert) {
        throw new Error(`Certificate not found: ${deployment.certificateId}`);
      }

      // Get private key from HSM
      const key = await this.hsmClient.listKeys().then(keys =>
        keys.find(k => k.id.toString('hex') === deployment.privateKeyId)
      );

      if (!key) {
        throw new Error(`Private key not found: ${deployment.privateKeyId}`);
      }

      // Create deployment directory if needed
      const deploymentDir = path.dirname(deployment.deploymentPath);
      if (!fs.existsSync(deploymentDir)) {
        fs.mkdirSync(deploymentDir, { recursive: true });
      }

      // Export certificate and chain
      const certPEM = this.certificateService.exportCertificatePEM(deployment.certificateId);
      
      fs.writeFileSync(
        `${deployment.deploymentPath}.crt`,
        certPEM
      );

      // Export chain if provided
      if (deployment.chainCertificateIds && deployment.chainCertificateIds.length > 0) {
        const chainPEM = deployment.chainCertificateIds
          .map(id => this.certificateService.exportCertificatePEM(id))
          .join('\n');
        fs.writeFileSync(
          `${deployment.deploymentPath}.chain.crt`,
          chainPEM
        );
      }

      // In production, private key would stay in HSM
      // and would be referenced by ID, not exported
      fs.writeFileSync(
        `${deployment.deploymentPath}.key.id`,
        deployment.privateKeyId
      );

      // Execute reload command if provided
      if (deployment.reloadCommand) {
        await this.executeCommand(deployment.reloadCommand);
      }

      // Restart service if provided
      if (deployment.restartService) {
        await this.executeCommand(`systemctl restart ${deployment.restartService}`);
      }

      console.log('[TLS-MANAGER] Certificate deployed successfully');

      // Update inventory
      const inventory = this.inventory.get(deployment.certificateId);
      if (inventory) {
        inventory.deployedServices.push(deployment.deploymentPath);
        inventory.lastChecked = new Date();
        this.inventory.set(deployment.certificateId, inventory);
      }

      // Emit deployment event
      this.emit('certificateDeployed', {
        type: 'deployed',
        certificateId: deployment.certificateId,
        commonName: cert.subject,
        message: `Certificate deployed to ${deployment.deploymentPath}`
      });

    } catch (error: any) {
      console.error('[TLS-MANAGER] Deployment failed:', error.message);
      this.emit('error', {
        type: 'error',
        certificateId: deployment.certificateId,
        commonName: '',
        message: `Deployment failed: ${error.message}`
      });
      throw error;
    }
  }

  /**
   * Creates an HTTPS server with TLS
   * 
   * @param serviceName - Service identifier
   * @param options - Server options
   * @returns TLS server instance
   */
  async createHTTPSServer(
    serviceName: string,
    options: {
      certificateId: string;
      keyId: string;
      requestCert?: boolean;
    }
  ): Promise<tls.Server> {
    console.log('[TLS-MANAGER] Creating HTTPS server for:', serviceName);

    const cert = this.certificateService.getCertificate(options.certificateId);
    if (!cert) {
      throw new Error(`Certificate not found: ${options.certificateId}`);
    }

    // Get key from HSM (would use HSM-based key handler in production)
    const keys = await this.hsmClient.listKeys();
    const key = keys.find(k => k.id.toString('hex') === options.keyId);

    if (!key) {
      throw new Error(`Key not found: ${options.keyId}`);
    }

    // Create TLS options
    const tlsOptions: tls.TlsOptions = {
      cert: this.certificateService.exportCertificatePEM(options.certificateId),
      // In production, would use HSM-based key provider
      // key would not be exported
      minVersion: this.config.minVersion as any,
      maxVersion: this.config.maxVersion as any,
      ciphers: this.config.ciphers,
      honorCipherOrder: this.config.honorCipherOrder,
      requestCert: options.requestCert || this.config.requestClientCert,
      rejectUnauthorized: this.config.rejectUnauthorized,
      SNICallback: (servername, callback) => {
        // In production, would select certificate based on servername
        callback(null, undefined);
      }
    };

    // Create server
    const server = tls.createServer(tlsOptions, (socket) => {
      console.log(`[TLS-MANAGER] Client connected: ${socket.authorized}`);
    });

    // Store server instance
    this.serverInstances.set(serviceName, server);

    console.log('[TLS-MANAGER] HTTPS server created');

    return server;
  }

  /**
   * Starts monitoring certificates
   */
  startMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
    }

    // Initial check
    this.checkCertificates();

    // Schedule periodic checks
    this.monitoringInterval = setInterval(() => {
      this.checkCertificates();
    }, this.monitoringConfig.checkIntervalMinutes * 60 * 1000);

    console.log('[TLS-MANAGER] Certificate monitoring started');
  }

  /**
   * Stops monitoring
   */
  stopMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
    console.log('[TLS-MANAGER] Certificate monitoring stopped');
  }

  /**
   * Checks all certificates for expiration
   */
  async checkCertificates(): Promise<void> {
    console.log('[TLS-MANAGER] Checking certificates');

    const certificates = this.certificateService.listCertificates();
    const now = new Date();

    for (const cert of certificates) {
      const daysUntilExpiry = Math.ceil(
        (cert.notAfter.getTime() - now.getTime()) / (1000 * 60 * 60 * 24)
      );

      // Update inventory
      let inventory = this.inventory.get(cert.id);
      if (!inventory) {
        inventory = await this.addToInventory(cert);
      }

      inventory.daysUntilExpiry = daysUntilExpiry;
      inventory.lastChecked = now;
      this.inventory.set(cert.id, inventory);

      // Check for expiration
      if (daysUntilExpiry <= 0) {
        this.emit('certificateExpired', {
          type: 'expired',
          certificateId: cert.id,
          commonName: cert.subject,
          daysUntilExpiry: 0,
          message: `Certificate has expired: ${cert.subject}`
        });
      } else if (daysUntilExpiry <= this.monitoringConfig.criticalThresholdDays) {
        this.emit('certificateCritical', {
          type: 'expiring',
          certificateId: cert.id,
          commonName: cert.subject,
          daysUntilExpiry,
          message: `Certificate expires in ${daysUntilExpiry} days (CRITICAL)`
        });

        // Auto-renew if enabled
        if (this.monitoringConfig.autoRenew) {
          await this.renewCertificate(cert.id);
        }
      } else if (daysUntilExpiry <= this.monitoringConfig.warningThresholdDays) {
        this.emit('certificateExpiring', {
          type: 'expiring',
          certificateId: cert.id,
          commonName: cert.subject,
          daysUntilExpiry,
          message: `Certificate expires in ${daysUntilExpiry} days (WARNING)`
        });
      }
    }

    console.log(`[TLS-MANAGER] Certificate check completed: ${certificates.length} certificates`);
  }

  /**
   * Renews a certificate
   */
  async renewCertificate(certificateId: string): Promise<CertificateInfo> {
    console.log('[TLS-MANAGER] Renewing certificate:', certificateId);

    const oldCert = this.certificateService.getCertificate(certificateId);
    if (!oldCert) {
      throw new Error(`Certificate not found: ${certificateId}`);
    }

    // Check if auto-renew is allowed
    if (!this.monitoringConfig.autoRenewEnabled) {
      throw new Error('Auto-renewal is not enabled');
    }

    // Create new certificate
    const newCert = await this.certificateService.renewCertificate(certificateId);

    // Update inventory
    const oldInventory = this.inventory.get(certificateId);
    if (oldInventory) {
      oldInventory.status = CertificateStatus.RENEWED;
      this.inventory.set(certificateId, oldInventory);
    }

    await this.addToInventory(newCert);

    // Emit renewal event
    this.emit('certificateRenewed', {
      type: 'renewed',
      certificateId: newCert.id,
      commonName: newCert.subject,
      daysUntilExpiry: Math.ceil(
        (newCert.notAfter.getTime() - Date.now()) / (1000 * 60 * 60 * 24)
      ),
      message: `Certificate renewed: ${newCert.subject}`
    });

    return newCert;
  }

  /**
   * Revokes a certificate
   */
  async revokeCertificate(
    certificateId: string,
    reason: string = 'KeyCompromise'
  ): Promise<void> {
    console.log('[TLS-MANAGER] Revoking certificate:', certificateId);

    await this.certificateService.revokeCertificate(certificateId, reason);

    // Update inventory
    const inventory = this.inventory.get(certificateId);
    if (inventory) {
      inventory.status = CertificateStatus.REVOKED;
      this.inventory.set(certificateId, inventory);
    }

    // Emit revocation event
    this.emit('certificateRevoked', {
      type: 'error',
      certificateId,
      commonName: inventory?.commonName || '',
      message: `Certificate revoked: ${reason}`
    });
  }

  /**
   * Gets certificate inventory
   */
  getInventory(): CertificateInventory[] {
    return Array.from(this.inventory.values());
  }

  /**
   * Gets expiring certificates
   */
  getExpiringCertificates(days: number = 30): CertificateInventory[] {
    return Array.from(this.inventory.values()).filter(
      cert => cert.daysUntilExpiry <= days && cert.status === CertificateStatus.ACTIVE
    );
  }

  /**
   * Registers a service endpoint
   */
  registerService(endpoint: ServiceEndpoint): void {
    this.services.set(endpoint.name, endpoint);
    console.log('[TLS-MANAGER] Service registered:', endpoint.name);
  }

  /**
   * Gets TLS configuration status
   */
  getStatus(): {
    config: TLSConfig;
    monitoring: MonitoringConfig;
    certificateCount: number;
    serviceCount: number;
    activeServers: number;
    expiringCount: number;
  } {
    return {
      config: this.config,
      monitoring: this.monitoringConfig,
      certificateCount: this.inventory.size,
      serviceCount: this.services.size,
      activeServers: this.serverInstances.size,
      expiringCount: this.getExpiringCertificates(
        this.monitoringConfig.warningThresholdDays
      ).length
    };
  }

  /**
   * Updates monitoring configuration
   */
  updateMonitoringConfig(config: Partial<MonitoringConfig>): void {
    this.monitoringConfig = {
      ...this.monitoringConfig,
      ...config
    };

    // Restart monitoring if interval changed
    if (this.monitoringInterval) {
      this.startMonitoring();
    }
  }

  /**
   * Stops all servers and cleans up
   */
  async shutdown(): Promise<void> {
    console.log('[TLS-MANAGER] Shutting down');

    // Stop monitoring
    this.stopMonitoring();

    // Close all server instances
    for (const [name, server] of this.serverInstances) {
      await new Promise<void>((resolve) => {
        server.close(() => resolve());
      });
      console.log('[TLS-MANAGER] Server stopped:', name);
    }

    this.serverInstances.clear();
    console.log('[TLS-MANAGER] TLS Certificate Manager shutdown complete');
  }

  // Private helper methods

  /**
   * Adds certificate to inventory
   */
  private async addToInventory(cert: CertificateInfo): Promise<CertificateInventory> {
    const inventory: CertificateInventory = {
      id: cert.id,
      commonName: this.extractCN(cert.subject),
      type: cert.type,
      status: cert.status,
      issuer: cert.issuer,
      serialNumber: cert.serialNumber,
      fingerprint: cert.fingerprint,
      notBefore: cert.notBefore,
      notAfter: cert.notAfter,
      daysUntilExpiry: Math.ceil(
        (cert.notAfter.getTime() - Date.now()) / (1000 * 60 * 60 * 24)
      ),
      deployedServices: [],
      lastChecked: new Date(),
      autoRenew: false
    };

    this.inventory.set(cert.id, inventory);
    return inventory;
  }

  /**
   * Extracts common name from subject
   */
  private extractCN(subject: string): string {
    const match = subject.match(/CN=([^,]+)/);
    return match ? match[1] : subject;
  }

  /**
   * Loads inventory from storage
   */
  private async loadInventory(): Promise<void> {
    // In production, would load from database
    console.log('[TLS-MANAGER] Loading certificate inventory');
  }

  /**
   * Executes a shell command
   */
  private async executeCommand(command: string): Promise<void> {
    try {
      const { exec } = require('child_process');
      await new Promise<void>((resolve, reject) => {
        exec(command, (error: any, stdout: string, stderr: string) => {
          if (error) {
            reject(error);
          } else {
            resolve();
          }
        });
      });
      console.log('[TLS-MANAGER] Command executed:', command);
    } catch (error: any) {
      console.error('[TLS-MANAGER] Command failed:', command, error.message);
      throw error;
    }
  }
}

// Export types
export {
  TLSConfig,
  CertificateDeployment,
  MonitoringConfig,
  CertificateInventory,
  ServiceEndpoint,
  TLSMonitoringEvent
};
