/**
 * Reports API Routes
 * 
 * Endpoints for report generation, history, and download.
 */

import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { DataSource } from 'typeorm';
import { getDataSource } from '../config/database';
import { ReportHistory, ReportStatus, ReportTrigger } from '../models/reportHistory.entity';
import { ReportSchedule, ScheduleStatus } from '../models/reportSchedule.entity';
import { ReportTemplate } from '../models/reportTemplate.entity';
import { s3StorageService } from '../services/s3Storage.service';
import { kafkaService } from '../services/kafka.service';
import { prometheusService } from '../services/prometheus.service';
import { ReportQueue } from '../queue/reportQueue.service';

export const reportRoutes = Router();
const reportQueue = new ReportQueue();

// Get all reports with filtering and pagination
reportRoutes.get('/', async (req: Request, res: Response) => {
  try {
    const {
      page = 1,
      limit = 20,
      status,
      reportType,
      entityId,
      format,
      startDate,
      endDate,
    } = req.query;

    const dataSource: DataSource = getDataSource();
    const reportRepo = dataSource.getRepository(ReportHistory);

    const queryBuilder = reportRepo.createQueryBuilder('report')
      .leftJoinAndSelect('report.schedule', 'schedule')
      .leftJoinAndSelect('report.template', 'template');

    if (status) {
      queryBuilder.andWhere('report.status = :status', { status });
    }
    if (reportType) {
      queryBuilder.andWhere('report.reportType = :reportType', { reportType });
    }
    if (entityId) {
      queryBuilder.andWhere('report.entityId = :entityId', { entityId });
    }
    if (format) {
      queryBuilder.andWhere('report.format = :format', { format });
    }
    if (startDate) {
      queryBuilder.andWhere('report.generatedAt >= :startDate', { startDate });
    }
    if (endDate) {
      queryBuilder.andWhere('report.generatedAt <= :endDate', { endDate });
    }

    queryBuilder
      .orderBy('report.createdAt', 'DESC')
      .skip((Number(page) - 1) * Number(limit))
      .take(Number(limit));

    const [data, total] = await queryBuilder.getManyAndCount();

    prometheusService.incrementApiCounter('GET', '/reports', 200);

    res.json({
      data,
      pagination: {
        page: Number(page),
        limit: Number(limit),
        total,
        pages: Math.ceil(total / Number(limit)),
      },
    });
  } catch (error: any) {
    console.error('Error fetching reports:', error);
    prometheusService.incrementApiCounter('GET', '/reports', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Generate a new report
reportRoutes.post('/generate', async (req: Request, res: Response) => {
  try {
    const {
      reportType,
      format = 'pdf',
      entityId,
      parameters = {},
      filterCriteria = {},
      scheduleId,
      periodStart,
      periodEnd,
      recipients,
    } = req.body;

    const dataSource: DataSource = getDataSource();

    // Validate report type exists
    const templateRepo = dataSource.getRepository(ReportTemplate);
    const template = await templateRepo.findOne({
      where: { reportType, isActive: true },
    });

    if (!template) {
      return res.status(400).json({
        status: 'error',
        message: `Invalid report type: ${reportType}`,
      });
    }

    // Create report history record
    const reportId = uuidv4();
    const reportHistory = dataSource.getRepository(ReportHistory).create({
      id: reportId,
      entityId,
      reportType,
      name: `${template.name} - ${new Date().toISOString()}`,
      format,
      status: ReportStatus.PENDING,
      trigger: scheduleId ? ReportTrigger.SCHEDULED : ReportTrigger.MANUAL,
      scheduleId,
      parameters,
      filterCriteria,
      reportPeriodStart: periodStart ? new Date(periodStart) : null,
      reportPeriodEnd: periodEnd ? new Date(periodEnd) : null,
      requestedBy: (req as any).user?.id || 'anonymous',
    });

    await dataSource.getRepository(ReportHistory).save(reportHistory);

    // Add to queue
    await reportQueue.addJob({
      reportId,
      reportType,
      format,
      parameters: {
        ...parameters,
        templateId: template.id,
        templateContent: template.templateContent,
        templateVariables: template.templateVariables,
        defaultParameters: template.defaultParameters,
      },
      entityId,
      requestedBy: (req as any).user?.id || 'anonymous',
      scheduleId,
      trigger: scheduleId ? 'scheduled' : 'manual',
    });

    prometheusService.incrementApiCounter('POST', '/reports/generate', 201);

    res.status(201).json({
      status: 'success',
      data: {
        reportId,
        status: 'queued',
        reportType,
        format,
        createdAt: reportHistory.createdAt,
      },
    });
  } catch (error: any) {
    console.error('Error generating report:', error);
    prometheusService.incrementApiCounter('POST', '/reports/generate', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Get single report details
reportRoutes.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();

    const report = await dataSource.getRepository(ReportHistory).findOne({
      where: { id },
      relations: ['schedule', 'template'],
    });

    if (!report) {
      prometheusService.incrementApiCounter('GET', '/reports/:id', 404);
      return res.status(404).json({
        status: 'error',
        message: 'Report not found',
      });
    }

    prometheusService.incrementApiCounter('GET', '/reports/:id', 200);

    res.json({
      data: report,
    });
  } catch (error: any) {
    console.error('Error fetching report:', error);
    prometheusService.incrementApiCounter('GET', '/reports/:id', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Download report
reportRoutes.get('/:id/download', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();
    const reportRepo = dataSource.getRepository(ReportHistory);

    const report = await reportRepo.findOne({ where: { id } });

    if (!report) {
      return res.status(404).json({
        status: 'error',
        message: 'Report not found',
      });
    }

    if (report.status !== ReportStatus.COMPLETED || !report.s3Key) {
      return res.status(400).json({
        status: 'error',
        message: 'Report not ready for download',
      });
    }

    // Verify file integrity if hash is stored
    if (report.fileHash) {
      const integrityCheck = await s3StorageService.verifyFileIntegrity(
        report.s3Key,
        report.fileHash,
      );

      if (!integrityCheck.valid) {
        console.error('Report integrity check failed:', integrityCheck.message);
        // Log but still allow download
      }
    }

    // Get signed URL
    const signedUrl = await s3StorageService.getSignedDownloadUrl(report.s3Key);

    // Update download tracking (chain of custody)
    const downloadLog = report.downloadLog || [];
    downloadLog.push({
      timestamp: new Date(),
      userId: (req as any).user?.id || 'anonymous',
      ipAddress: req.ip || req.socket.remoteAddress,
      userAgent: req.get('User-Agent') || 'unknown',
    });

    report.downloadCount = (report.downloadCount || 0) + 1;
    report.lastDownloadedAt = new Date();
    report.downloadLog = downloadLog;

    await reportRepo.save(report);

    prometheusService.incrementDownloadCounter(report.reportType, report.format);

    // Publish download event
    await kafkaService.publishEvent('report.downloaded', {
      reportId: id,
      reportType: report.reportType,
      downloadedBy: (req as any).user?.id || 'anonymous',
    });

    // Redirect to signed URL
    res.redirect(signedUrl);
  } catch (error: any) {
    console.error('Error downloading report:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Verify report integrity
reportRoutes.get('/:id/verify', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();
    const reportRepo = dataSource.getRepository(ReportHistory);

    const report = await reportRepo.findOne({ where: { id } });

    if (!report) {
      return res.status(404).json({
        status: 'error',
        message: 'Report not found',
      });
    }

    if (!report.s3Key || !report.fileHash) {
      return res.status(400).json({
        status: 'error',
        message: 'Report integrity verification not available',
      });
    }

    const integrityCheck = await s3StorageService.verifyFileIntegrity(
      report.s3Key,
      report.fileHash,
    );

    res.json({
      data: {
        reportId: id,
        storedHash: report.fileHash,
        actualHash: integrityCheck.actualHash,
        valid: integrityCheck.valid,
        message: integrityCheck.message,
        verifiedAt: new Date().toISOString(),
        chainOfCustody: {
          downloadCount: report.downloadCount,
          lastDownloadedAt: report.lastDownloadedAt,
          downloadLog: report.downloadLog?.length || 0,
        },
        wormCompliance: {
          retentionUntil: report.retentionUntil,
          retentionMode: report.retentionMode,
          legalHold: report.legalHold,
        },
      },
    });
  } catch (error: any) {
    console.error('Error verifying report:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Get report generation status
reportRoutes.get('/:id/status', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const status = await reportQueue.getJobStatus(id);

    if (!status) {
      return res.status(404).json({
        status: 'error',
        message: 'Report job not found',
      });
    }

    res.json({
      data: status,
    });
  } catch (error: any) {
    console.error('Error getting status:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Cancel report generation
reportRoutes.post('/:id/cancel', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const cancelled = await reportQueue.cancelJob(id);

    if (cancelled) {
      // Update status in database
      const dataSource: DataSource = getDataSource();
      await dataSource.getRepository(ReportHistory).update(id, {
        status: ReportStatus.CANCELLED,
      });

      res.json({
        status: 'success',
        message: 'Report generation cancelled',
      });
    } else {
      res.status(400).json({
        status: 'error',
        message: 'Unable to cancel report - may already be processing',
      });
    }
  } catch (error: any) {
    console.error('Error cancelling report:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});
