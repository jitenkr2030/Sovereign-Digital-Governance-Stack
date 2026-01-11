import {
  Injectable,
  NotFoundException,
  BadRequestException,
  Logger,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { ConfigService } from '@nestjs/config';
import * as crypto from 'crypto';
import { S3Client, PutObjectCommand, GetObjectCommand, DeleteObjectCommand } from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { Document, DocumentStatus, DocumentType } from './entities/document.entity';

@Injectable()
export class DocumentsService {
  private readonly logger = new Logger(DocumentsService.name);
  private s3Client: S3Client;
  private bucket: string;

  constructor(
    @InjectRepository(Document)
    private documentRepository: Repository<Document>,
    private configService: ConfigService,
  ) {
    this.s3Client = new S3Client({
      endpoint: this.configService.get('S3_ENDPOINT', 'http://localhost:9000'),
      region: this.configService.get('S3_REGION', 'us-east-1'),
      credentials: {
        accessKeyId: this.configService.get('S3_ACCESS_KEY', 'minioadmin'),
        secretAccessKey: this.configService.get('S3_SECRET_KEY', 'minioadmin'),
      },
    });
    this.bucket = this.configService.get('S3_BUCKET', 'csic-documents');
  }

  async uploadDocument(
    applicationId: string,
    file: {
      originalname: string;
      mimetype: string;
      buffer: Buffer;
      size: number;
    },
    documentType: DocumentType,
    uploadedBy: string,
    uploadedByName: string,
  ): Promise<Document> {
    // Calculate checksum for integrity verification
    const checksum = this.calculateChecksum(file.buffer);
    const s3Key = this.generateS3Key(applicationId, documentType, file.originalname);

    // Upload to S3
    await this.s3Client.send(
      new PutObjectCommand({
        Bucket: this.bucket,
        Key: s3Key,
        Body: file.buffer,
        ContentType: file.mimetype,
        Metadata: {
          checksum,
          originalName: file.originalname,
          uploadedBy,
        },
      }),
    );

    // Create document record
    const document = this.documentRepository.create({
      applicationId,
      fileName: s3Key,
      originalName: file.originalname,
      mimeType: file.mimetype,
      fileSize: file.size,
      s3Key,
      s3Bucket: this.bucket,
      documentType,
      checksum,
      checksumAlgorithm: 'sha256',
      status: DocumentStatus.UPLOADED,
      uploadedBy,
      uploadedByName,
    });

    const saved = await this.documentRepository.save(document);

    this.logger.log(`Document uploaded: ${saved.id} (${file.originalname})`);

    return saved;
  }

  async getDocument(id: string): Promise<Document> {
    const document = await this.documentRepository.findOne({
      where: { id },
      relations: ['application'],
    });

    if (!document) {
      throw new NotFoundException(`Document with ID ${id} not found`);
    }

    return document;
  }

  async getSignedDownloadUrl(id: string): Promise<string> {
    const document = await this.getDocument(id);

    const command = new GetObjectCommand({
      Bucket: document.s3Bucket,
      Key: document.s3Key,
    });

    const url = await getSignedUrl(this.s3Client, command, {
      expiresIn: this.configService.get('S3_SIGNED_URL_EXPIRY', 3600),
    });

    return url;
  }

  async verifyDocument(id: string, verifiedBy: string, verifiedByName: string, notes?: string): Promise<Document> {
    const document = await this.getDocument(id);

    // Re-calculate checksum to verify integrity
    const currentChecksum = await this.calculateStoredChecksum(document.s3Key);

    if (currentChecksum !== document.checksum) {
      document.status = DocumentStatus.QUARANTINED;
      const quarantineDays = this.configService.get('DOCUMENT_QUARANTINE_DAYS', 7);
      document.quarantineUntil = new Date();
      document.quarantineUntil.setDate(document.quarantineUntil.getDate() + quarantineDays);
      await this.documentRepository.save(document);

      this.logger.warn(`Document ${id} failed checksum verification - quarantined`);
      throw new BadRequestException('Document integrity check failed - quarantined for review');
    }

    document.status = DocumentStatus.VERIFIED;
    document.verifiedAt = new Date();
    document.verifiedBy = verifiedBy;
    document.verifiedByName = verifiedByName;
    document.verificationNotes = notes;

    return this.documentRepository.save(document);
  }

  async rejectDocument(id: string, reason: string, rejectedBy: string): Promise<Document> {
    const document = await this.getDocument(id);

    document.status = DocumentStatus.REJECTED;
    document.rejectedAt = new Date();
    document.rejectionReason = reason;

    const saved = await this.documentRepository.save(document);

    this.logger.log(`Document ${id} rejected: ${reason}`);

    return saved;
  }

  async getDocumentsByApplication(applicationId: string): Promise<Document[]> {
    return this.documentRepository.find({
      where: { applicationId },
      order: { createdAt: 'DESC' },
    });
  }

  async getPendingVerification(): Promise<Document[]> {
    return this.documentRepository.find({
      where: { status: DocumentStatus.PENDING_VERIFICATION },
      order: { createdAt: 'ASC' },
    });
  }

  async deleteDocument(id: string): Promise<void> {
    const document = await this.getDocument(id);

    // Delete from S3
    try {
      await this.s3Client.send(
        new DeleteObjectCommand({
          Bucket: document.s3Bucket,
          Key: document.s3Key,
        }),
      );
    } catch (error) {
      this.logger.error(`Failed to delete S3 object: ${document.s3Key}`, error);
    }

    // Delete record
    await this.documentRepository.remove(document);

    this.logger.log(`Document deleted: ${id}`);
  }

  async verifyIntegrity(id: string): Promise<{ valid: boolean; message: string }> {
    const document = await this.getDocument(id);

    const currentChecksum = await this.calculateStoredChecksum(document.s3Key);

    if (currentChecksum === document.checksum) {
      return { valid: true, message: 'Document integrity verified' };
    }

    return { valid: false, message: 'Document integrity check failed - checksum mismatch' };
  }

  private calculateChecksum(buffer: Buffer): string {
    return crypto.createHash('sha256').update(buffer).digest('hex');
  }

  private async calculateStoredChecksum(s3Key: string): Promise<string> {
    try {
      const response = await this.s3Client.send(
        new GetObjectCommand({
          Bucket: this.bucket,
          Key: s3Key,
        }),
      );

      const chunks: Buffer[] = [];
      for await (const chunk of response.Body as any) {
        chunks.push(Buffer.from(chunk));
      }
      const buffer = Buffer.concat(chunks);

      return this.calculateChecksum(buffer);
    } catch (error) {
      throw new Error(`Failed to calculate checksum: ${error.message}`);
    }
  }

  private generateS3Key(applicationId: string, documentType: DocumentType, originalName: string): string {
    const timestamp = Date.now();
    const sanitizedName = originalName.replace(/[^a-zA-Z0-9.-]/g, '_');
    return `applications/${applicationId}/${documentType}/${timestamp}-${sanitizedName}`;
  }
}
