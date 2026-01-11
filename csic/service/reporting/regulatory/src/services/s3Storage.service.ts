/**
 * S3 Storage Service
 * 
 * Handles file storage operations with S3-compatible storage.
 */

import {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
  DeleteObjectCommand,
  HeadObjectCommand,
  CopyObjectCommand,
  PutObjectLockConfigurationCommand,
} from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { v4 as uuidv4 } from 'uuid';
import * as crypto from 'crypto';
import { config } from '../config/config';

export interface UploadResult {
  key: string;
  bucket: string;
  size: number;
  hash: string;
  url: string;
}

export interface WORMConfig {
  retentionDays: number;
  retentionMode: 'GOVERNANCE' | 'COMPLIANCE';
  legalHold: boolean;
}

export class S3StorageService {
  private client: S3Client;
  private bucket: string;
  private defaultWormConfig: WORMConfig;

  constructor() {
    const s3Config = config.storage.s3;
    this.bucket = s3Config.bucket;
    this.defaultWormConfig = {
      retentionDays: config.storage.worm.retention_days,
      retentionMode: config.storage.worm.retention_mode as 'GOVERNANCE' | 'COMPLIANCE',
      legalHold: config.storage.worm.legal_hold_enabled,
    };

    this.client = new S3Client({
      endpoint: s3Config.endpoint,
      region: s3Config.region,
      credentials: {
        accessKeyId: s3Config.access_key,
        secretAccessKey: s3Config.secret_key,
      },
      forcePathStyle: true,
    });
  }

  /**
   * Upload a file with optional WORM protection
   */
  async uploadFile(
    fileBuffer: Buffer,
    fileName: string,
    contentType: string,
    wormConfig?: WORMConfig,
  ): Promise<UploadResult> {
    const key = this.generateKey(fileName);
    const hash = this.calculateHash(fileBuffer);

    // Upload to S3
    const command = new PutObjectCommand({
      Bucket: this.bucket,
      Key: key,
      Body: fileBuffer,
      ContentType: contentType,
      Metadata: {
        fileHash: hash,
        uploadedAt: new Date().toISOString(),
      },
    });

    await this.client.send(command);

    // Apply WORM configuration if enabled
    if (config.storage.worm.enabled || wormConfig) {
      const retentionConfig = wormConfig || this.defaultWormConfig;
      await this.applyWORMConfiguration(key, retentionConfig);
    }

    // Get signed URL for download
    const url = await this.getSignedDownloadUrl(key);

    return {
      key,
      bucket: this.bucket,
      size: fileBuffer.length,
      hash,
      url,
    };
  }

  /**
   * Apply WORM (Write Once Read Many) configuration to an object
   */
  async applyWORMConfiguration(key: string, wormConfig: WORMConfig): Promise<void> {
    const retentionDate = new Date();
    retentionDate.setDate(retentionDate.getDate() + wormConfig.retentionDays);

    const command = new PutObjectLockConfigurationCommand({
      Bucket: this.bucket,
      Key: key,
      ObjectLockConfiguration: {
        ObjectLockEnabled: 'Enabled',
        Rule: {
          ObjectLockRetainUntilDate: retentionDate,
          Mode: wormConfig.retentionMode,
        },
      },
    });

    await this.client.send(command);
  }

  /**
   * Get a signed URL for downloading a file
   */
  async getSignedDownloadUrl(key: string, expiresIn: number = 3600): Promise<string> {
    const command = new GetObjectCommand({
      Bucket: this.bucket,
      Key: key,
    });

    return getSignedUrl(this.client, command, { expiresIn });
  }

  /**
   * Get file metadata
   */
  async getFileMetadata(key: string): Promise<{
    size: number;
    contentType: string;
    hash: string;
    lastModified: Date;
    wormInfo?: {
      retentionUntil: Date;
      mode: string;
      legalHold: boolean;
    };
  }> {
    const command = new HeadObjectCommand({
      Bucket: this.bucket,
      Key: key,
    });

    const response = await this.client.send(command);

    return {
      size: response.ContentLength || 0,
      contentType: response.ContentType || 'application/octet-stream',
      hash: response.Metadata?.filehash || '',
      lastModified: response.LastModified || new Date(),
    };
  }

  /**
   * Verify file integrity using stored hash
   */
  async verifyFileIntegrity(key: string, expectedHash: string): Promise<{
    valid: boolean;
    actualHash: string;
    message: string;
  }> {
    const command = new GetObjectCommand({
      Bucket: this.bucket,
      Key: key,
    });

    const response = await this.client.send(command);

    // Read the file content to calculate hash
    const chunks: Buffer[] = [];
    for await (const chunk of response.Body as any) {
      chunks.push(Buffer.from(chunk));
    }
    const fileBuffer = Buffer.concat(chunks);
    const actualHash = this.calculateHash(fileBuffer);

    return {
      valid: actualHash === expectedHash,
      actualHash,
      message: actualHash === expectedHash
        ? 'File integrity verified'
        : 'File integrity check failed - file may have been tampered with',
    };
  }

  /**
   * Delete a file (only allowed for non-WORM objects)
   */
  async deleteFile(key: string): Promise<void> {
    const command = new DeleteObjectCommand({
      Bucket: this.bucket,
      Key: key,
    });

    await this.client.send(command);
  }

  /**
   * Copy a file to a new location
   */
  async copyFile(sourceKey: string, destinationKey: string): Promise<void> {
    const command = new CopyObjectCommand({
      Bucket: this.bucket,
      CopySource: `${this.bucket}/${sourceKey}`,
      Key: destinationKey,
    });

    await this.client.send(command);
  }

  /**
   * Generate a unique S3 key for a file
   */
  private generateKey(fileName: string): string {
    const date = new Date().toISOString().split('T')[0];
    const uniqueId = uuidv4().substring(0, 8);
    const sanitizedName = fileName.replace(/[^a-zA-Z0-9.-]/g, '_');
    return `reports/${date}/${uniqueId}-${sanitizedName}`;
  }

  /**
   * Calculate SHA-256 hash of file content
   */
  private calculateHash(buffer: Buffer): string {
    return crypto.createHash('sha256').update(buffer).digest('hex');
  }
}

// Export singleton instance
export const s3StorageService = new S3StorageService();
