/**
 * Prometheus Metrics Service
 * 
 * Collects and exposes Prometheus metrics for monitoring.
 */

import * as promClient from 'prom-client';

export class PrometheusService {
  private register: promClient.Registry;
  private counters: Map<string, promClient.Counter>;
  private histograms: Map<string, promClient.Histogram>;
  private gauges: Map<string, promClient.Gauge>;

  constructor() {
    this.register = new promClient.Registry();
    this.counters = new Map();
    this.histograms = new Map();
    this.gauges = new Map();

    // Add default metrics
    promClient.collectDefaultMetrics({ register: this.register });

    // Register custom metrics
    this.initializeMetrics();
  }

  private initializeMetrics(): void {
    // Report generation counters
    this.counters.set('reports_generated_total', new promClient.Counter({
      name: 'csic_reports_generated_total',
      help: 'Total number of reports generated',
      labelNames: ['report_type', 'format', 'status'],
      registers: [this.register],
    }));

    this.counters.set('report_generation_errors', new promClient.Counter({
      name: 'csic_report_generation_errors_total',
      help: 'Total number of report generation errors',
      labelNames: ['report_type', 'error_type'],
      registers: [this.register],
    }));

    this.counters.set('report_downloads_total', new promClient.Counter({
      name: 'csic_report_downloads_total',
      help: 'Total number of report downloads',
      labelNames: ['report_type', 'format'],
      registers: [this.register],
    }));

    this.counters.set('api_requests_total', new promClient.Counter({
      name: 'csic_reporting_api_requests_total',
      help: 'Total API requests',
      labelNames: ['method', 'endpoint', 'status_code'],
      registers: [this.register],
    }));

    // Histograms
    this.histograms.set('report_generation_duration', new promClient.Histogram({
      name: 'csic_report_generation_duration_seconds',
      help: 'Report generation duration in seconds',
      labelNames: ['report_type', 'format'],
      buckets: [1, 5, 10, 30, 60, 120, 300],
      registers: [this.register],
    }));

    this.histograms.set('api_request_duration', new promClient.Histogram({
      name: 'csic_reporting_api_request_duration_seconds',
      help: 'API request duration in seconds',
      labelNames: ['method', 'endpoint'],
      buckets: [0.01, 0.05, 0.1, 0.5, 1, 5, 10],
      registers: [this.register],
    }));

    // Gauges
    this.gauges.set('reports_in_queue', new promClient.Gauge({
      name: 'csic_reports_in_queue',
      help: 'Number of reports currently in the queue',
      registers: [this.register],
    }));

    this.gauges.set('storage_used_bytes', new promClient.Gauge({
      name: 'csic_reporting_storage_used_bytes',
      help: 'Total storage used by reports in bytes',
      registers: [this.register],
    }));
  }

  incrementReportCounter(reportType: string, format: string, status: string): void {
    const counter = this.counters.get('reports_generated_total');
    if (counter) {
      counter.inc({ report_type: reportType, format, status });
    }
  }

  incrementErrorCounter(errorType: string, reportType?: string): void {
    const counter = this.counters.get('report_generation_errors');
    if (counter) {
      counter.inc({ report_type: reportType || 'unknown', error_type: errorType });
    }
  }

  incrementDownloadCounter(reportType: string, format: string): void {
    const counter = this.counters.get('report_downloads_total');
    if (counter) {
      counter.inc({ report_type: reportType, format });
    }
  }

  incrementApiCounter(method: string, endpoint: string, statusCode: number): void {
    const counter = this.counters.get('api_requests_total');
    if (counter) {
      counter.inc({ method, endpoint, status_code: statusCode.toString() });
    }
  }

  observeReportDuration(reportType: string, format: string, durationSeconds: number): void {
    const histogram = this.histograms.get('report_generation_duration');
    if (histogram) {
      histogram.observe({ report_type: reportType, format }, durationSeconds);
    }
  }

  observeApiDuration(method: string, endpoint: string, durationSeconds: number): void {
    const histogram = this.histograms.get('api_request_duration');
    if (histogram) {
      histogram.observe({ method, endpoint }, durationSeconds);
    }
  }

  setQueueSize(size: number): void {
    const gauge = this.gauges.get('reports_in_queue');
    if (gauge) {
      gauge.set(size);
    }
  }

  setStorageUsed(bytes: number): void {
    const gauge = this.gauges.get('storage_used_bytes');
    if (gauge) {
      gauge.set(bytes);
    }
  }

  getContentType(): string {
    return this.register.contentType;
  }

  async getMetrics(): Promise<string> {
    return this.register.metrics();
  }
}

// Export singleton instance
export const prometheusService = new PrometheusService();
