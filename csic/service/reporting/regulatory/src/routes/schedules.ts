/**
 * Schedules API Routes
 * 
 * Endpoints for managing report schedules.
 */

import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { DataSource } from 'typeorm';
import { getDataSource } from '../config/database';
import { ReportSchedule, ScheduleStatus, ScheduleTrigger } from '../models/reportSchedule.entity';
import { ReportHistory, ReportTrigger } from '../models/reportHistory.entity';
import { prometheusService } from '../services/prometheus.service';

export const scheduleRoutes = Router();

// Get all schedules
scheduleRoutes.get('/', async (req: Request, res: Response) => {
  try {
    const {
      page = 1,
      limit = 20,
      status,
      reportType,
      entityId,
    } = req.query;

    const dataSource: DataSource = getDataSource();
    const scheduleRepo = dataSource.getRepository(ReportSchedule);

    const queryBuilder = scheduleRepo.createQueryBuilder('schedule')
      .leftJoinAndSelect('schedule.template', 'template');

    if (status) {
      queryBuilder.andWhere('schedule.status = :status', { status });
    }
    if (reportType) {
      queryBuilder.andWhere('schedule.reportType = :reportType', { reportType });
    }
    if (entityId) {
      queryBuilder.andWhere('schedule.entityId = :entityId', { entityId });
    }

    queryBuilder
      .orderBy('schedule.createdAt', 'DESC')
      .skip((Number(page) - 1) * Number(limit))
      .take(Number(limit));

    const [data, total] = await queryBuilder.getManyAndCount();

    prometheusService.incrementApiCounter('GET', '/schedules', 200);

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
    console.error('Error fetching schedules:', error);
    prometheusService.incrementApiCounter('GET', '/schedules', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Create new schedule
scheduleRoutes.post('/', async (req: Request, res: Response) => {
  try {
    const {
      entityId,
      name,
      description,
      reportType,
      templateId,
      trigger = 'CRON',
      cronExpression,
      parameters,
      recipients,
      filterCriteria,
      defaultFormat = 'pdf',
      includeAttachments = true,
    } = req.body;

    // Validate required fields
    if (!name || !reportType || !templateId) {
      return res.status(400).json({
        status: 'error',
        message: 'Missing required fields: name, reportType, templateId',
      });
    }

    if (trigger === 'CRON' && !cronExpression) {
      return res.status(400).json({
        status: 'error',
        message: 'cronExpression is required for CRON trigger',
      });
    }

    const dataSource: DataSource = getDataSource();
    const scheduleRepo = dataSource.getRepository(ReportSchedule);

    const schedule = scheduleRepo.create({
      id: uuidv4(),
      entityId,
      name,
      description,
      reportType,
      templateId,
      trigger: trigger as ScheduleTrigger,
      cronExpression,
      parameters,
      recipients,
      filterCriteria,
      defaultFormat,
      includeAttachments,
      status: ScheduleStatus.ACTIVE,
      createdBy: (req as any).user?.id || 'anonymous',
    });

    await scheduleRepo.save(schedule);

    prometheusService.incrementApiCounter('POST', '/schedules', 201);

    res.status(201).json({
      status: 'success',
      data: schedule,
    });
  } catch (error: any) {
    console.error('Error creating schedule:', error);
    prometheusService.incrementApiCounter('POST', '/schedules', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Get single schedule
scheduleRoutes.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();

    const schedule = await dataSource.getRepository(ReportSchedule).findOne({
      where: { id },
      relations: ['template', 'history'],
    });

    if (!schedule) {
      prometheusService.incrementApiCounter('GET', '/schedules/:id', 404);
      return res.status(404).json({
        status: 'error',
        message: 'Schedule not found',
      });
    }

    // Get schedule statistics
    const historyStats = await dataSource.getRepository(ReportHistory)
      .createQueryBuilder('history')
      .select('history.status', 'status')
      .addSelect('COUNT(*)', 'count')
      .where('history.scheduleId = :id', { id })
      .groupBy('history.status')
      .getRawMany();

    const stats = {
      totalRuns: schedule.runCount,
      lastRunAt: schedule.lastRunAt,
      nextRunAt: schedule.nextRunAt,
      historyByStatus: historyStats.reduce((acc, h) => {
        acc[h.status] = parseInt(h.count);
        return acc;
      }, {}),
    };

    prometheusService.incrementApiCounter('GET', '/schedules/:id', 200);

    res.json({
      data: {
        ...schedule,
        statistics: stats,
      },
    });
  } catch (error: any) {
    console.error('Error fetching schedule:', error);
    prometheusService.incrementApiCounter('GET', '/schedules/:id', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Update schedule
scheduleRoutes.patch('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const updates = req.body;

    const dataSource: DataSource = getDataSource();
    const scheduleRepo = dataSource.getRepository(ReportSchedule);

    const schedule = await scheduleRepo.findOne({ where: { id } });

    if (!schedule) {
      return res.status(404).json({
        status: 'error',
        message: 'Schedule not found',
      });
    }

    // Only allow updating certain fields
    const allowedUpdates = [
      'name', 'description', 'cronExpression', 'parameters',
      'recipients', 'filterCriteria', 'defaultFormat', 'includeAttachments', 'status',
    ];

    for (const key of Object.keys(updates)) {
      if (allowedUpdates.includes(key)) {
        (schedule as any)[key] = updates[key];
      }
    }

    schedule.updatedBy = (req as any).user?.id || 'anonymous';
    await scheduleRepo.save(schedule);

    res.json({
      status: 'success',
      data: schedule,
    });
  } catch (error: any) {
    console.error('Error updating schedule:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Delete schedule
scheduleRoutes.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();
    const scheduleRepo = dataSource.getRepository(ReportSchedule);

    const schedule = await scheduleRepo.findOne({ where: { id } });

    if (!schedule) {
      return res.status(404).json({
        status: 'error',
        message: 'Schedule not found',
      });
    }

    // Soft delete by disabling
    schedule.status = ScheduleStatus.DISABLED;
    await scheduleRepo.save(schedule);

    res.json({
      status: 'success',
      message: 'Schedule disabled',
    });
  } catch (error: any) {
    console.error('Error deleting schedule:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Trigger schedule manually
scheduleRoutes.post('/:id/trigger', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();
    const scheduleRepo = dataSource.getRepository(ReportSchedule);

    const schedule = await scheduleRepo.findOne({
      where: { id },
      relations: ['template'],
    });

    if (!schedule) {
      return res.status(404).json({
        status: 'error',
        message: 'Schedule not found',
      });
    }

    if (schedule.status !== ScheduleStatus.ACTIVE) {
      return res.status(400).json({
        status: 'error',
        message: 'Schedule is not active',
      });
    }

    // Create a manual report generation
    const reportId = uuidv4();
    const reportHistoryRepo = dataSource.getRepository(ReportHistory);
    
    const report = reportHistoryRepo.create({
      id: reportId,
      scheduleId: schedule.id,
      entityId: schedule.entityId,
      reportType: schedule.reportType,
      name: `${schedule.name} - Manual Trigger`,
      format: schedule.defaultFormat,
      status: 'PENDING',
      trigger: ReportTrigger.SCHEDULED,
      parameters: schedule.parameters,
      filterCriteria: schedule.filterCriteria,
      requestedBy: (req as any).user?.id || 'anonymous',
    });

    await reportHistoryRepo.save(report);

    // Update schedule statistics
    schedule.runCount += 1;
    schedule.lastRunAt = new Date();
    await scheduleRepo.save(schedule);

    res.json({
      status: 'success',
      data: {
        reportId,
        scheduleId: schedule.id,
        triggeredAt: new Date().toISOString(),
      },
    });
  } catch (error: any) {
    console.error('Error triggering schedule:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Get schedule history
scheduleRoutes.get('/:id/history', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const { page = 1, limit = 20 } = req.query;

    const dataSource: DataSource = getDataSource();
    const historyRepo = dataSource.getRepository(ReportHistory);

    const [data, total] = await historyRepo.findAndCount({
      where: { scheduleId: id },
      order: { createdAt: 'DESC' },
      skip: (Number(page) - 1) * Number(limit),
      take: Number(limit),
    });

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
    console.error('Error fetching schedule history:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});
