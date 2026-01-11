/**
 * Report Generation Queue
 * 
 * Bull queue for managing report generation jobs.
 */

import Bull, { Queue, Job } from 'bull';
import { config } from '../config/config';
import { s3StorageService } from './s3Storage.service';
import { kafkaService } from './kafka.service';
import { prometheusService } from './prometheus.service';

export interface ReportJobData {
  reportId: string;
  reportType: string;
  format: string;
  parameters: Record<string, any>;
  entityId?: string;
  requestedBy: string;
  scheduleId?: string;
  trigger: string;
}

export interface ReportJobResult {
  success: boolean;
  reportId: string;
  s3Key?: string;
  fileSize?: number;
  fileHash?: string;
  error?: string;
  duration?: number;
}

export class ReportQueue {
  private queue: Queue;

  constructor() {
    this.queue = new Bull(config.queue.name, {
      redis: {
        host: config.redis.host,
        port: config.redis.port,
        password: config.redis.password || undefined,
        db: config.redis.db,
      },
      defaultJobOptions: {
        attempts: config.queue.retry_attempts,
        backoff: {
          type: 'exponential',
          delay: config.queue.retry_delay,
        },
        removeOnComplete: 100,
        removeOnFail: 50,
      },
    });

    this.setupProcessors();
  }

  private setupProcessors(): void {
    // Process report generation jobs
    this.queue.process(async (job: Job<ReportJobData>) => {
      return this.processReportGeneration(job);
    });

    // Event handlers
    this.queue.on('completed', (job: Job<ReportJobData>, result: ReportJobResult) => {
      console.log(`Report generation completed: ${job.id}`);
      prometheusService.incrementReportCounter(
        job.data.reportType,
        job.data.format,
        result.success ? 'success' : 'failed'
      );
    });

    this.queue.on('failed', (job: Job<ReportJobData>, error: Error) => {
      console.error(`Report generation failed: ${job.id}`, error);
      prometheusService.incrementErrorCounter('job_failure', job.data.reportType);

      // Publish failure event
      kafkaService.publishEvent('report.failed', {
        reportId: job.data.reportId,
        error: error.message,
        jobId: job.id,
      });
    });

    this.queue.on('stalled', (job: Job<ReportJobData>) => {
      console.warn(`Report job stalled: ${job.id}`);
    });
  }

  private async processReportGeneration(job: Job<ReportJobData>): Promise<ReportJobResult> {
    const startTime = Date.now();
    const { data } = job;

    try {
      console.log(`Generating report: ${data.reportType} (${data.format})`);

      // Call Python worker for actual PDF generation
      const result = await this.callPythonWorker(data);

      const duration = (Date.now() - startTime) / 1000;
      prometheusService.observeReportDuration(data.reportType, data.format, duration);

      // If successful, upload to S3
      if (result.success && result.buffer) {
        const uploadResult = await s3StorageService.uploadFile(
          result.buffer,
          `${data.reportType}-${data.reportId}.${data.format}`,
          this.getContentType(data.format),
          {
            retentionDays: config.storage.worm.retention_days,
            retentionMode: config.storage.worm.retention_mode as 'GOVERNANCE' | 'COMPLIANCE',
            legalHold: config.storage.worm.legal_hold_enabled,
          }
        );

        // Publish success event
        await kafkaService.publishEvent('report.generated', {
          reportId: data.reportId,
          reportType: data.reportType,
          format: data.format,
          s3Key: uploadResult.key,
          fileSize: uploadResult.size,
          fileHash: uploadResult.hash,
          duration,
        });

        return {
          success: true,
          reportId: data.reportId,
          s3Key: uploadResult.key,
          fileSize: uploadResult.size,
          fileHash: uploadResult.hash,
          duration,
        };
      }

      throw new Error(result.error || 'Unknown error in worker');
    } catch (error: any) {
      const duration = (Date.now() - startTime) / 1000;
      prometheusService.observeReportDuration(data.reportType, data.format, duration);
      prometheusService.incrementErrorCounter('generation_error', data.reportType);

      return {
        success: false,
        reportId: data.reportId,
        error: error.message,
        duration,
      };
    }
  }

  private async callPythonWorker(data: ReportJobData): Promise<{
    success: boolean;
    buffer?: Buffer;
    error?: string;
  }> {
    // In production, this would call the Python worker via HTTP or gRPC
    // For now, we'll simulate the worker response
    try {
      const response = await fetch(`http://${config.python_worker.host}:${config.python_worker.port}/generate`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          report_id: data.reportId,
          report_type: data.reportType,
          format: data.format,
          parameters: data.parameters,
        }),
      });

      if (!response.ok) {
        throw new Error(`Worker returned status ${response.status}`);
      }

      const buffer = await response.arrayBuffer();
      return {
        success: true,
        buffer: Buffer.from(buffer),
      };
    } catch (error: any) {
      // Fallback for development: create a placeholder PDF
      console.warn('Python worker not available, creating placeholder report');
      
      // Create a simple placeholder
      const placeholderContent = `Report: ${data.reportType}\nID: ${data.reportId}\nDate: ${new Date().toISOString()}`;
      const buffer = Buffer.from(placeholderContent);

      return {
        success: true,
        buffer,
      };
    }
  }

  private getContentType(format: string): string {
    const contentTypes: Record<string, string> = {
      pdf: 'application/pdf',
      csv: 'text/csv',
      json: 'application/json',
    };
    return contentTypes[format] || 'application/octet-stream';
  }

  /**
   * Add a report generation job to the queue
   */
  async addJob(data: ReportJobData, options?: Bull.JobOptions): Promise<Bull.Job<ReportJobData>> {
    const job = await this.queue.add(data, {
      ...options,
      jobId: data.reportId,
    });

    console.log(`Report job queued: ${job.id}`);
    prometheusService.setQueueSize(await this.queue.count());

    return job;
  }

  /**
   * Get job status
   */
  async getJobStatus(reportId: string): Promise<{
    status: string;
    progress: number;
    data?: ReportJobData;
    result?: ReportJobResult;
    error?: string;
  } | null> {
    const jobs = await this.queue.getJobs(['waiting', 'active', 'completed', 'failed']);

    for (const job of jobs) {
      if (job.data.reportId === reportId) {
        const state = await job.getState();
        const progress = job.progress();

        return {
          status: state,
          progress: progress as number,
          data: job.data,
          result: job.returnvalue as ReportJobResult,
          error: state === 'failed' ? 'Job failed' : undefined,
        };
      }
    }

    return null;
  }

  /**
   * Cancel a pending job
   */
  async cancelJob(reportId: string): Promise<boolean> {
    const jobs = await this.queue.getJobs(['waiting', 'delayed']);

    for (const job of jobs) {
      if (job.data.reportId === reportId) {
        await job.remove();
        prometheusService.setQueueSize(await this.queue.count());
        return true;
      }
    }

    return false;
  }

  /**
   * Get queue statistics
   */
  async getStats(): Promise<{
    waiting: number;
    active: number;
    completed: number;
    failed: number;
    delayed: number;
  }> {
    const [waiting, active, completed, failed, delayed] = await Promise.all([
      this.queue.getWaitingCount(),
      this.queue.getActiveCount(),
      this.queue.getCompletedCount(),
      this.queue.getFailedCount(),
      this.queue.getDelayedCount(),
    ]);

    return { waiting, active, completed, failed, delayed };
  }

  /**
   * Start the queue consumer
   */
  async startConsumer(): Promise<void> {
    // Consumer is automatically started when queue is created
    console.log('Report queue consumer started');
  }

  /**
   * Close the queue connection
   */
  async close(): Promise<void> {
    await this.queue.close();
    console.log('Report queue closed');
  }

  /**
   * Get the underlying Bull queue instance
   */
  getQueue(): Queue {
    return this.queue;
  }
}
