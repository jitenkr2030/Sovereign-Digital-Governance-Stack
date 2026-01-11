import { Injectable } from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository, Between } from 'typeorm';
import { Application, ApplicationStatus, ApplicationPriority } from '../applications/entities/application.entity';
import { License, LicenseStatus } from '../licenses/entities/license.entity';
import { LicenseType } from '../applications/entities/license-type.entity';

@Injectable()
export class DashboardService {
  constructor(
    @InjectRepository(Application)
    private applicationRepository: Repository<Application>,
    @InjectRepository(License)
    private licenseRepository: Repository<License>,
    @InjectRepository(LicenseType)
    private licenseTypeRepository: Repository<LicenseType>,
  ) {}

  async getOverviewStats(): Promise<any> {
    const [
      applicationStats,
      licenseStats,
      licenseTypeStats,
      recentActivity,
    ] = await Promise.all([
      this.getApplicationStats(),
      this.getLicenseStats(),
      this.getLicenseTypeStats(),
      this.getRecentActivity(),
    ]);

    return {
      applications: applicationStats,
      licenses: licenseStats,
      byLicenseType: licenseTypeStats,
      recentActivity,
      generatedAt: new Date().toISOString(),
    };
  }

  async getApplicationStats(): Promise<any> {
    const byStatus = await this.applicationRepository
      .createQueryBuilder('application')
      .select('application.status', 'status')
      .addSelect('COUNT(*)', 'count')
      .groupBy('application.status')
      .getRawMany();

    const byPriority = await this.applicationRepository
      .createQueryBuilder('application')
      .select('application.priority', 'priority')
      .addSelect('COUNT(*)', 'count')
      .where('application.status IN (:...statuses)', {
        statuses: [ApplicationStatus.SUBMITTED, ApplicationStatus.UNDER_REVIEW],
      })
      .groupBy('application.priority')
      .getRawMany();

    const avgReviewTime = await this.calculateAverageReviewTime();

    return {
      total: await this.applicationRepository.count(),
      byStatus: byStatus.reduce(
        (acc, s) => ({ ...acc, [s.status]: parseInt(s.count) }),
        {},
      ),
      byPriority: byPriority.reduce(
        (acc, p) => ({ ...acc, [p.priority]: parseInt(p.count) }),
        {},
      ),
      pendingReview: await this.applicationRepository.count({
        where: {
          status: ApplicationStatus.UNDER_REVIEW,
        },
      }),
      averageReviewTimeDays: avgReviewTime,
    };
  }

  async getLicenseStats(): Promise<any> {
    const byStatus = await this.licenseRepository
      .createQueryBuilder('license')
      .select('license.status', 'status')
      .addSelect('COUNT(*)', 'count')
      .groupBy('license.status')
      .getRawMany();

    const expiringWithin30Days = await this.getExpiringCount(30);
    const expiringWithin60Days = await this.getExpiringCount(60);
    const expiringWithin90Days = await this.getExpiringCount(90);

    return {
      total: await this.licenseRepository.count(),
      byStatus: byStatus.reduce(
        (acc, s) => ({ ...acc, [s.status]: parseInt(s.count) }),
        {},
      ),
      active: await this.licenseRepository.count({
        where: { status: LicenseStatus.ACTIVE },
      }),
      expiring: {
        within30Days: expiringWithin30Days,
        within60Days: expiringWithin60Days,
        within90Days: expiringWithin90Days,
      },
      suspended: await this.licenseRepository.count({
        where: { status: LicenseStatus.SUSPENDED },
      }),
      revoked: await this.licenseRepository.count({
        where: { status: LicenseStatus.REVOKED },
      }),
    };
  }

  async getLicenseTypeStats(): Promise<any> {
    const stats = await this.licenseRepository
      .createQueryBuilder('license')
      .select('license.licenseTypeName', 'type')
      .addSelect('COUNT(*)', 'total')
      .addSelect('SUM(CASE WHEN license.status = :active THEN 1 ELSE 0 END)', 'active')
      .setParameter('active', LicenseStatus.ACTIVE)
      .groupBy('license.licenseTypeName')
      .getRawMany();

    return stats.map(s => ({
      type: s.type,
      total: parseInt(s.total),
      active: parseInt(s.active),
    }));
  }

  async getRecentActivity(): Promise<any[]> {
    const thirtyDaysAgo = new Date();
    thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);

    const recentApplications = await this.applicationRepository.find({
      where: {
        createdAt: MoreThan(thirtyDaysAgo),
      },
      order: { createdAt: 'DESC' },
      take: 10,
      select: ['id', 'entityName', 'status', 'createdAt', 'licenseType'],
    });

    const recentLicenses = await this.licenseRepository.find({
      where: {
        createdAt: MoreThan(thirtyDaysAgo),
      },
      order: { createdAt: 'DESC' },
      take: 10,
      select: ['id', 'entityName', 'licenseNumber', 'status', 'createdAt'],
    });

    return [
      ...recentApplications.map(a => ({
        type: 'application',
        action: a.status,
        entity: a.entityName,
        timestamp: a.createdAt,
      })),
      ...recentLicenses.map(l => ({
        type: 'license',
        action: l.status,
        entity: l.entityName,
        timestamp: l.createdAt,
      })),
    ]
      .sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime())
      .slice(0, 10);
  }

  async getQueueMetrics(): Promise<any> {
    const now = new Date();

    const submittedToday = await this.applicationRepository.count({
      where: {
        status: ApplicationStatus.SUBMITTED,
        submittedAt: Between(
          new Date(now.setHours(0, 0, 0, 0)),
          new Date(now.setHours(23, 59, 59, 999)),
        ),
      },
    });

    const reviewedToday = await this.applicationRepository.count({
      where: {
        status: ApplicationStatus.UNDER_REVIEW,
        reviewStartedAt: Between(
          new Date(now.setHours(0, 0, 0, 0)),
          new Date(now.setHours(23, 59, 59, 999)),
        ),
      },
    });

    const decidedToday = await this.applicationRepository.count({
      where: {
        decisionAt: Between(
          new Date(now.setHours(0, 0, 0, 0)),
          new Date(now.setHours(23, 59, 59, 999)),
        ),
      },
    });

    const licensesIssuedToday = await this.licenseRepository.count({
      where: {
        createdAt: Between(
          new Date(now.setHours(0, 0, 0, 0)),
          new Date(now.setHours(23, 59, 59, 999)),
        ),
      },
    });

    return {
      submittedToday,
      reviewedToday,
      decidedToday,
      licensesIssuedToday,
      queueDepth: await this.applicationRepository.count({
        where: {
          status: ApplicationStatus.UNDER_REVIEW,
        },
      }),
    };
  }

  async getComplianceMetrics(): Promise<any> {
    const activeLicenses = await this.licenseRepository.find({
      where: { status: LicenseStatus.ACTIVE },
    });

    const expiringCount = await this.getExpiringCount(30);
    const suspendedCount = await this.licenseRepository.count({
      where: { status: LicenseStatus.SUSPENDED },
    });

    const renewalRate = await this.calculateRenewalRate();

    return {
      totalActiveLicenses: activeLicenses.length,
      complianceRate: activeLicenses.length > 0
        ? ((activeLicenses.length - suspendedCount) / activeLicenses.length) * 100
        : 100,
      expiringIn30Days: expiringCount,
      suspensionRate: activeLicenses.length > 0
        ? (suspendedCount / activeLicenses.length) * 100
        : 0,
      renewalRate,
    };
  }

  private async getExpiringCount(days: number): Promise<number> {
    const today = new Date();
    const futureDate = new Date(today);
    futureDate.setDate(futureDate.getDate() + days);

    return this.licenseRepository.count({
      where: {
        status: LicenseStatus.ACTIVE,
        expiryDate: Between(today, futureDate),
      },
    });
  }

  private async calculateAverageReviewTime(): Promise<number> {
    const completedApplications = await this.applicationRepository
      .createQueryBuilder('application')
      .select(
        'AVG(EXTRACT(EPOCH FROM (application.decisionAt - application.submittedAt)) / 86400)',
        'avgDays',
      )
      .where('application.decisionAt IS NOT NULL')
      .getRawOne();

    return parseFloat(completedApplications?.avgDays || 0).toFixed(1);
  }

  private async calculateRenewalRate(): Promise<number> {
    const renewedLicenses = await this.licenseRepository.count({
      where: {
        renewedAt: MoreThan(new Date(Date.now() - 365 * 24 * 60 * 60 * 1000)),
      },
    });

    const expiringLicenses = await this.licenseRepository.count({
      where: {
        expiryDate: Between(
          new Date(Date.now() - 365 * 24 * 60 * 60 * 1000),
          new Date(),
        ),
      },
    });

    return expiringLicenses > 0
      ? (renewedLicenses / expiringLicenses) * 100
      : 100;
  }
}
