import { Injectable, Logger } from '@nestjs/common';
import { Cron, CronExpression } from '@nestjs/schedule';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, Between, LessThan, MoreThan } from 'typeorm';
import { License, LicenseStatus } from '../licenses/entities/license.entity';
import { KafkaService } from '../kafka/kafka.service';
import { ConfigService } from '@nestjs/config';

@Injectable()
export class RenewalsService {
  private readonly logger = new Logger(RenewalsService.name);
  private readonly alertDays: number[];

  constructor(
    @InjectRepository(License)
    private licenseRepository: Repository<License>,
    private kafkaService: KafkaService,
    private configService: ConfigService,
  ) {
    this.alertDays = this.configService.get<number[]>('RENEWAL_ALERT_DAYS', [90, 60, 30, 14, 7, 1]);
  }

  @Cron(CronExpression.EVERY_DAY_AT_9AM)
  async processRenewalAlerts(): Promise<void> {
    this.logger.log('Processing renewal alerts...');

    const today = new Date();

    for (const days of this.alertDays) {
      const targetDate = new Date(today);
      targetDate.setDate(targetDate.getDate() + days);

      const startOfDay = new Date(targetDate);
      startOfDay.setHours(0, 0, 0, 0);
      const endOfDay = new Date(targetDate);
      endOfDay.setHours(23, 59, 59, 999);

      const expiringLicenses = await this.licenseRepository.find({
        where: {
          status: LicenseStatus.ACTIVE,
          expiryDate: Between(startOfDay, endOfDay),
        },
      });

      for (const license of expiringLicenses) {
        await this.sendRenewalAlert(license, days);
      }

      this.logger.log(`Found ${expiringLicenses.length} licenses expiring in ${days} days`);
    }

    this.logger.log('Renewal alerts processing complete');
  }

  @Cron(CronExpression.EVERY_HOUR)
  async processExpiredLicenses(): Promise<void> {
    this.logger.log('Processing expired licenses...');

    const yesterday = new Date();
    yesterday.setDate(yesterday.getDate() - 1);
    yesterday.setHours(23, 59, 59, 999);

    const expiredLicenses = await this.licenseRepository.find({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: LessThan(yesterday),
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
        expiryDate: license.expiryDate,
        expiredAt: new Date(),
      });

      this.logger.log(`License ${license.licenseNumber} marked as expired`);
    }

    if (expiredLicenses.length > 0) {
      this.logger.log(`Processed ${expiredLicenses.length} expired licenses`);
    }
  }

  @Cron(CronExpression.EVERY_DAY_AT_8AM)
  async openRenewalWindows(): Promise<void> {
    this.logger.log('Opening renewal windows...');

    const today = new Date();

    const licensesToOpen = await this.licenseRepository.find({
      where: {
        status: LicenseStatus.ACTIVE,
        renewalWindowStart: LessThan(today),
      },
    });

    // These licenses are already in ACTIVE status but now within renewal window
    // We could update them to PENDING_RENEWAL status if needed
    this.logger.log(`${licensesToOpen.length} licenses have open renewal windows`);
  }

  private async sendRenewalAlert(license: License, daysUntilExpiry: number): Promise<void> {
    await this.kafkaService.publishLicenseEvent('license.expiring', {
      licenseId: license.id,
      entityId: license.entityId,
      entityName: license.entityName,
      licenseNumber: license.licenseNumber,
      licenseType: license.licenseTypeName,
      expiryDate: license.expiryDate,
      daysUntilExpiry,
      renewalWindowStart: license.renewalWindowStart,
    });

    this.logger.log(
      `Sent renewal alert for license ${license.licenseNumber}: ${daysUntilExpiry} days until expiry`,
    );
  }

  async getUpcomingRenewals(days: number = 90): Promise<License[]> {
    const today = new Date();
    const futureDate = new Date(today);
    futureDate.setDate(futureDate.getDate() + days);

    return this.licenseRepository.find({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: Between(today, futureDate),
      },
      order: { expiryDate: 'ASC' },
    });
  }

  async getRenewalsSummary(): Promise<{
    totalExpiring: number;
    byTimeframe: Record<string, number>;
    alertsSentToday: number;
  }> {
    const today = new Date();
    const byTimeframe: Record<string, number> = {};

    for (const days of this.alertDays) {
      const targetDate = new Date(today);
      targetDate.setDate(targetDate.getDate() + days);

      const startOfDay = new Date(targetDate);
      startOfDay.setHours(0, 0, 0, 0);
      const endOfDay = new Date(targetDate);
      endOfDay.setHours(23, 59, 59, 999);

      const count = await this.licenseRepository.count({
        where: {
          status: LicenseStatus.ACTIVE,
          expiryDate: Between(startOfDay, endOfDay),
        },
      });

      byTimeframe[`${days}_days`] = count;
    }

    const totalExpiring = Object.values(byTimeframe).reduce((sum, count) => sum + count, 0);

    return {
      totalExpiring,
      byTimeframe,
      alertsSentToday: 0, // Would need separate tracking
    };
  }
}
