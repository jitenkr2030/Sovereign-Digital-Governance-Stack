import { Injectable, NotFoundException, BadRequestException } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, LessThan, MoreThan, Between } from 'typeorm';
import { v4 as uuidv4 } from 'uuid';
import { License, LicenseStatus } from './entities/license.entity';
import { Application, ApplicationStatus } from '../applications/entities/application.entity';
import { KafkaService } from '../kafka/kafka.service';
import { WorkflowService } from '../workflow/workflow.service';

@Injectable()
export class LicensesService {
  constructor(
    @InjectRepository(License)
    private licenseRepository: Repository<License>,
    @InjectRepository(Application)
    private applicationRepository: Repository<Application>,
    private kafkaService: KafkaService,
    private workflowService: WorkflowService,
  ) {}

  async issueLicense(applicationId: string, issuerId: string, issuerName: string): Promise<License> {
    const application = await this.applicationRepository.findOne({
      where: { id: applicationId },
      relations: ['licenseType'],
    });

    if (!application) {
      throw new NotFoundException('Application not found');
    }

    if (application.status !== ApplicationStatus.APPROVED) {
      throw new BadRequestException('Application must be approved before issuing license');
    }

    // Generate license number
    const licenseNumber = this.generateLicenseNumber(application.licenseType);

    // Calculate dates
    const issueDate = new Date();
    const validityMonths = application.licenseType?.validityMonths || 12;
    const expiryDate = new Date(issueDate);
    expiryDate.setMonth(expiryDate.getMonth() + validityMonths);

    const renewalWindowStart = new Date(expiryDate);
    renewalWindowStart.setDate(renewalWindowStart.getDate() - 90); // 90 days before expiry

    // Create license
    const license = this.licenseRepository.create({
      applicationId: application.id,
      entityId: application.entityId,
      entityName: application.entityName,
      entityType: application.entityType,
      licenseNumber,
      licenseTypeId: application.licenseTypeId,
      licenseTypeName: application.licenseType?.name,
      status: LicenseStatus.ACTIVE,
      issueDate,
      expiryDate,
      renewalWindowStart,
      scope: this.getDefaultScope(application.licenseType?.code),
      conditions: this.getDefaultConditions(application.licenseType?.code),
    });

    const saved = await this.licenseRepository.save(license);

    // Update application status
    application.status = ApplicationStatus.RENEWAL_WINDOW;
    await this.applicationRepository.save(application);

    // Publish Kafka event
    await this.kafkaService.publishLicenseEvent('license.issued', {
      licenseId: saved.id,
      applicationId: application.id,
      entityId: application.entityId,
      entityName: application.entityName,
      licenseNumber,
      licenseType: application.licenseType?.name,
      issueDate,
      expiryDate,
      issuedBy: issuerName,
    });

    return saved;
  }

  async findAll(filter: {
    status?: LicenseStatus;
    entityId?: string;
    licenseTypeId?: string;
    expiringWithinDays?: number;
    page?: number;
    limit?: number;
  }): Promise<{ data: License[]; total: number; page: number; limit: number }> {
    const { page = 1, limit = 20, status, entityId, licenseTypeId, expiringWithinDays } = filter;

    const queryBuilder = this.licenseRepository
      .createQueryBuilder('license');

    if (status) {
      queryBuilder.andWhere('license.status = :status', { status });
    }

    if (entityId) {
      queryBuilder.andWhere('license.entityId = :entityId', { entityId });
    }

    if (licenseTypeId) {
      queryBuilder.andWhere('license.licenseTypeId = :licenseTypeId', { licenseTypeId });
    }

    if (expiringWithinDays) {
      const futureDate = new Date();
      futureDate.setDate(futureDate.getDate() + expiringWithinDays);
      const now = new Date();
      queryBuilder.andWhere('license.expiryDate BETWEEN :now AND :futureDate', {
        now,
        futureDate,
      });
    }

    queryBuilder
      .orderBy('license.expiryDate', 'ASC')
      .skip((page - 1) * limit)
      .take(limit);

    const [data, total] = await queryBuilder.getManyAndCount();

    return { data, total, page, limit };
  }

  async findOne(id: string): Promise<License> {
    const license = await this.licenseRepository.findOne({
      where: { id },
    });

    if (!license) {
      throw new NotFoundException(`License with ID ${id} not found`);
    }

    return license;
  }

  async findByLicenseNumber(licenseNumber: string): Promise<License> {
    const license = await this.licenseRepository.findOne({
      where: { licenseNumber },
    });

    if (!license) {
      throw new NotFoundException(`License ${licenseNumber} not found`);
    }

    return license;
  }

  async suspendLicense(id: string, reason: string, suspendedBy: string): Promise<License> {
    const license = await this.findOne(id);

    if (!this.workflowService.canTransitionLicense(license.status, LicenseStatus.SUSPENDED)) {
      throw new BadRequestException(`Cannot suspend license with status ${license.status}`);
    }

    license.status = LicenseStatus.SUSPENDED;
    license.suspendedAt = new Date();
    license.suspendedReason = reason;
    license.suspendedBy = suspendedBy;

    const saved = await this.licenseRepository.save(license);

    await this.kafkaService.publishLicenseEvent('license.suspended', {
      licenseId: saved.id,
      entityId: saved.entityId,
      entityName: saved.entityName,
      licenseNumber: saved.licenseNumber,
      suspendedAt: saved.suspendedAt,
      reason,
      suspendedBy,
    });

    return saved;
  }

  async revokeLicense(id: string, reason: string, revokedBy: string): Promise<License> {
    const license = await this.findOne(id);

    if (!this.workflowService.canTransitionLicense(license.status, LicenseStatus.REVOKED)) {
      throw new BadRequestException(`Cannot revoke license with status ${license.status}`);
    }

    license.status = LicenseStatus.REVOKED;
    license.revokedAt = new Date();
    license.revokedReason = reason;
    license.revokedBy = revokedBy;

    const saved = await this.licenseRepository.save(license);

    await this.kafkaService.publishLicenseEvent('license.revoked', {
      licenseId: saved.id,
      entityId: saved.entityId,
      entityName: saved.entityName,
      licenseNumber: saved.licenseNumber,
      revokedAt: saved.revokedAt,
      reason,
      revokedBy,
    });

    return saved;
  }

  async renewLicense(id: string): Promise<License> {
    const license = await this.findOne(id);

    if (license.status !== LicenseStatus.PENDING_RENEWAL && license.status !== LicenseStatus.ACTIVE) {
      throw new BadRequestException('License must be active or pending renewal to renew');
    }

    const oldExpiryDate = license.expiryDate;

    // Calculate new dates
    const issueDate = new Date();
    const expiryDate = new Date(issueDate);
    expiryDate.setMonth(expiryDate.getMonth() + (license.licenseTypeName?.includes('Exchange') ? 12 : 24));

    license.status = LicenseStatus.ACTIVE;
    license.issueDate = issueDate;
    license.expiryDate = expiryDate;
    license.renewalWindowStart = new Date(expiryDate);
    license.renewalWindowStart.setDate(license.renewalWindowStart.getDate() - 90);
    license.renewedAt = new Date();

    const saved = await this.licenseRepository.save(license);

    await this.kafkaService.publishLicenseEvent('license.renewed', {
      licenseId: saved.id,
      entityId: saved.entityId,
      entityName: saved.entityName,
      licenseNumber: saved.licenseNumber,
      oldExpiryDate,
      newExpiryDate: saved.expiryDate,
      renewedAt: saved.renewedAt,
    });

    return saved;
  }

  async getExpiringLicenses(days: number): Promise<License[]> {
    const futureDate = new Date();
    futureDate.setDate(futureDate.getDate() + days);
    const now = new Date();

    return this.licenseRepository.find({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: Between(now, futureDate),
      },
      order: { expiryDate: 'ASC' },
    });
  }

  async processExpiredLicenses(): Promise<void> {
    const expiredLicenses = await this.licenseRepository.find({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: LessThan(new Date()),
      },
    });

    for (const license of expiredLicenses) {
      license.status = LicenseStatus.EXPIRED;
      await this.licenseRepository.save(license);

      await this.kafkaService.publishLicenseEvent('license.expired', {
        licenseId: license.id,
        entityId: license.entityId,
        entityName: license.entityName,
        licenseNumber: license.licenseNumber,
        expiredAt: new Date(),
        expiryDate: license.expiryDate,
      });
    }

    if (expiredLicenses.length > 0) {
      console.log(`Processed ${expiredLicenses.length} expired licenses`);
    }
  }

  async getStatistics(): Promise<any> {
    const statusStats = await this.licenseRepository
      .createQueryBuilder('license')
      .select('license.status', 'status')
      .addSelect('COUNT(*)', 'count')
      .groupBy('license.status')
      .getRawMany();

    const expiringCount = await this.licenseRepository.count({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: MoreThan(new Date()),
      },
    });

    return {
      byStatus: statusStats.reduce((acc, s) => ({ ...acc, [s.status]: parseInt(s.count) }), {}),
      totalActive: expiringCount,
      totalLicenses: await this.licenseRepository.count(),
    };
  }

  private generateLicenseNumber(licenseTypeCode: string): string {
    const prefix = licenseTypeCode?.toUpperCase().substring(0, 3) || 'LIC';
    const year = new Date().getFullYear();
    const uuid = uuidv4().substring(0, 8).toUpperCase();
    return `${prefix}-${year}-${uuid}`;
  }

  private getDefaultScope(licenseTypeCode: string): string[] {
    const scopes: Record<string, string[]> = {
      'vasp': ['Virtual Asset Custody', 'Virtual Asset Administration', 'Virtual Asset Transfer'],
      'casp': ['Crypto Asset Custody', 'Crypto Asset Trading'],
      'exchange': ['Spot Trading', 'Margin Trading', 'Derivatives Trading'],
      'wallet': ['Custodial Wallet Services', 'Hot Wallet Management'],
      'mining': ['Mining Pool Operation', 'Hashrate Trading'],
    };

    return scopes[licenseTypeCode] || ['General Operations'];
  }

  private getDefaultConditions(licenseTypeCode: string): any[] {
    return [
      {
        id: uuidv4(),
        title: 'AML/CFT Compliance',
        description: 'Maintain effective AML/CFT policies and procedures',
        isMet: true,
      },
      {
        id: uuidv4(),
        title: 'Quarterly Reporting',
        description: 'Submit quarterly activity reports to regulator',
        isMet: false,
        dueDate: this.getQuarterlyDueDate(),
      },
      {
        id: uuidv4(),
        title: 'Annual Audit',
        description: 'Complete annual security and financial audit',
        isMet: false,
        dueDate: this.getAnnualDueDate(),
      },
    ];
  }

  private getQuarterlyDueDate(): Date {
    const now = new Date();
    const quarter = Math.floor(now.getMonth() / 3);
    const nextQuarter = new Date(now.getFullYear(), (quarter + 1) * 3, 1);
    return nextQuarter;
  }

  private getAnnualDueDate(): Date {
    const nextYear = new Date();
    nextYear.setFullYear(nextYear.getFullYear() + 1);
    return nextYear;
  }
}
