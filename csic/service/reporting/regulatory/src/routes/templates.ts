/**
 * Templates API Routes
 * 
 * Endpoints for managing report templates.
 */

import { Router, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { DataSource } from 'typeorm';
import { getDataSource } from '../config/database';
import { ReportTemplate, TemplateFormat } from '../models/reportTemplate.entity';
import { prometheusService } from '../services/prometheus.service';

export const templateRoutes = Router();

// Get all templates
templateRoutes.get('/', async (req: Request, res: Response) => {
  try {
    const {
      page = 1,
      limit = 20,
      reportType,
      format,
      isActive,
    } = req.query;

    const dataSource: DataSource = getDataSource();
    const templateRepo = dataSource.getRepository(ReportTemplate);

    const queryBuilder = templateRepo.createQueryBuilder('template');

    if (reportType) {
      queryBuilder.andWhere('template.reportType = :reportType', { reportType });
    }
    if (format) {
      queryBuilder.andWhere('template.format = :format', { format });
    }
    if (isActive !== undefined) {
      queryBuilder.andWhere('template.isActive = :isActive', { isActive: isActive === 'true' });
    }

    queryBuilder
      .orderBy('template.createdAt', 'DESC')
      .skip((Number(page) - 1) * Number(limit))
      .take(Number(limit));

    const [data, total] = await queryBuilder.getManyAndCount();

    prometheusService.incrementApiCounter('GET', '/templates', 200);

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
    console.error('Error fetching templates:', error);
    prometheusService.incrementApiCounter('GET', '/templates', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Create new template
templateRoutes.post('/', async (req: Request, res: Response) => {
  try {
    const {
      name,
      description,
      reportType,
      format = 'pdf',
      templateContent,
      headerTemplate,
      footerTemplate,
      templateVariables,
      defaultParameters,
      styling,
    } = req.body;

    // Validate required fields
    if (!name || !reportType || !templateContent) {
      return res.status(400).json({
        status: 'error',
        message: 'Missing required fields: name, reportType, templateContent',
      });
    }

    const dataSource: DataSource = getDataSource();
    const templateRepo = dataSource.getRepository(ReportTemplate);

    // Check for duplicate
    const existing = await templateRepo.findOne({
      where: { reportType, isActive: true },
    });

    if (existing) {
      return res.status(409).json({
        status: 'error',
        message: 'Active template already exists for this report type',
      });
    }

    const template = templateRepo.create({
      id: uuidv4(),
      name,
      description,
      reportType,
      format: format as TemplateFormat,
      templateContent,
      headerTemplate,
      footerTemplate,
      templateVariables,
      defaultParameters,
      styling,
      isActive: true,
      createdBy: (req as any).user?.id || 'anonymous',
    });

    await templateRepo.save(template);

    prometheusService.incrementApiCounter('POST', '/templates', 201);

    res.status(201).json({
      status: 'success',
      data: template,
    });
  } catch (error: any) {
    console.error('Error creating template:', error);
    prometheusService.incrementApiCounter('POST', '/templates', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Get single template
templateRoutes.get('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();

    const template = await dataSource.getRepository(ReportTemplate).findOne({
      where: { id },
    });

    if (!template) {
      prometheusService.incrementApiCounter('GET', '/templates/:id', 404);
      return res.status(404).json({
        status: 'error',
        message: 'Template not found',
      });
    }

    prometheusService.incrementApiCounter('GET', '/templates/:id', 200);

    res.json({
      data: template,
    });
  } catch (error: any) {
    console.error('Error fetching template:', error);
    prometheusService.incrementApiCounter('GET', '/templates/:id', 500);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Update template
templateRoutes.put('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const updates = req.body;

    const dataSource: DataSource = getDataSource();
    const templateRepo = dataSource.getRepository(ReportTemplate);

    const template = await templateRepo.findOne({ where: { id } });

    if (!template) {
      return res.status(404).json({
        status: 'error',
        message: 'Template not found',
      });
    }

    // Update fields
    const allowedUpdates = [
      'name', 'description', 'templateContent', 'headerTemplate',
      'footerTemplate', 'templateVariables', 'defaultParameters', 'styling', 'isActive',
    ];

    for (const key of Object.keys(updates)) {
      if (allowedUpdates.includes(key)) {
        (template as any)[key] = updates[key];
      }
    }

    // Increment version
    template.version += 1;
    template.updatedBy = (req as any).user?.id || 'anonymous';

    await templateRepo.save(template);

    res.json({
      status: 'success',
      data: template,
    });
  } catch (error: any) {
    console.error('Error updating template:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Delete template (soft delete)
templateRoutes.delete('/:id', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const dataSource: DataSource = getDataSource();
    const templateRepo = dataSource.getRepository(ReportTemplate);

    const template = await templateRepo.findOne({ where: { id } });

    if (!template) {
      return res.status(404).json({
        status: 'error',
        message: 'Template not found',
      });
    }

    // Soft delete
    template.isActive = false;
    await templateRepo.save(template);

    res.json({
      status: 'success',
      message: 'Template deactivated',
    });
  } catch (error: any) {
    console.error('Error deleting template:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});

// Preview template
templateRoutes.post('/:id/preview', async (req: Request, res: Response) => {
  try {
    const { id } = req.params;
    const { data } = req.body;

    const dataSource: DataSource = getDataSource();
    const template = await dataSource.getRepository(ReportTemplate).findOne({
      where: { id },
    });

    if (!template) {
      return res.status(404).json({
        status: 'error',
        message: 'Template not found',
      });
    }

    // Simple Handlebars-like replacement for preview
    let preview = template.templateContent;
    if (data) {
      for (const [key, value] of Object.entries(data)) {
        preview = preview.replace(new RegExp(`{{${key}}}`, 'g'), String(value));
      }
    }

    res.json({
      data: {
        preview,
        variables: template.templateVariables,
      },
    });
  } catch (error: any) {
    console.error('Error previewing template:', error);
    res.status(500).json({
      status: 'error',
      message: error.message,
    });
  }
});
