import api from './api';
import { Report, ReportTemplate, ReportSchedule, ReportParameters, PaginatedResponse } from '../types';

interface ReportGenerateParams {
  title: string;
  type: string;
  format: 'pdf' | 'csv' | 'xlsx' | 'json';
  parameters: ReportParameters;
  entityId?: string;
  entityType?: string;
}

interface ReportListParams {
  type?: string[];
  status?: string[];
  startDate?: string;
  endDate?: string;
  page?: number;
  limit?: number;
}

interface ScheduleCreateParams {
  name: string;
  reportType: string;
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'annually';
  cronExpression?: string;
  parameters: ReportParameters;
  recipients: string[];
}

export const reportingService = {
  // Reports
  async getReports(params: ReportListParams = {}): Promise<PaginatedResponse<Report>> {
    const queryParams = new URLSearchParams();
    queryParams.append('page', params.page?.toString() || '1');
    queryParams.append('limit', params.limit?.toString() || '20');

    if (params.type?.length) queryParams.append('type', params.type.join(','));
    if (params.status?.length) queryParams.append('status', params.status.join(','));
    if (params.startDate) queryParams.append('startDate', params.startDate);
    if (params.endDate) queryParams.append('endDate', params.endDate);

    return api.get<PaginatedResponse<Report>>(`/reports?${queryParams}`);
  },

  async getReport(id: string): Promise<Report> {
    return api.get<Report>(`/reports/${id}`);
  },

  async generateReport(params: ReportGenerateParams): Promise<Report> {
    return api.post<Report>('/reports/generate', params);
  },

  async cancelReport(id: string): Promise<Report> {
    return api.post<Report>(`/reports/${id}/cancel`);
  },

  async retryReport(id: string): Promise<Report> {
    return api.post<Report>(`/reports/${id}/retry`);
  },

  async deleteReport(id: string): Promise<void> {
    return api.delete(`/reports/${id}`);
  },

  async downloadReport(id: string, filename?: string): Promise<void> {
    return api.download(`/reports/${id}/download`, filename || `report-${id}.pdf`);
  },

  async getReportStatus(id: string): Promise<{
    status: string;
    progress: number;
    message?: string;
  }> {
    return api.get(`/reports/${id}/status`);
  },

  // Report Templates
  async getTemplates(): Promise<ReportTemplate[]> {
    return api.get<ReportTemplate[]>('/reports/templates');
  },

  async getTemplate(id: string): Promise<ReportTemplate> {
    return api.get<ReportTemplate>(`/reports/templates/${id}`);
  },

  async createTemplate(data: Partial<ReportTemplate>): Promise<ReportTemplate> {
    return api.post<ReportTemplate>('/reports/templates', data);
  },

  async updateTemplate(id: string, data: Partial<ReportTemplate>): Promise<ReportTemplate> {
    return api.put<ReportTemplate>(`/reports/templates/${id}`, data);
  },

  async deleteTemplate(id: string): Promise<void> {
    return api.delete(`/reports/templates/${id}`);
  },

  async duplicateTemplate(id: string, newName: string): Promise<ReportTemplate> {
    return api.post<ReportTemplate>(`/reports/templates/${id}/duplicate`, { name: newName });
  },

  // Schedules
  async getSchedules(): Promise<ReportSchedule[]> {
    return api.get<ReportSchedule[]>('/reports/schedules');
  },

  async getSchedule(id: string): Promise<ReportSchedule> {
    return api.get<ReportSchedule>(`/reports/schedules/${id}`);
  },

  async createSchedule(data: ScheduleCreateParams): Promise<ReportSchedule> {
    return api.post<ReportSchedule>('/reports/schedules', data);
  },

  async updateSchedule(id: string, data: Partial<ReportSchedule>): Promise<ReportSchedule> {
    return api.put<ReportSchedule>(`/reports/schedules/${id}`, data);
  },

  async deleteSchedule(id: string): Promise<void> {
    return api.delete(`/reports/schedules/${id}`);
  },

  async toggleSchedule(id: string, isActive: boolean): Promise<ReportSchedule> {
    return api.post<ReportSchedule>(`/reports/schedules/${id}/toggle`, { isActive });
  },

  async triggerSchedule(id: string): Promise<Report> {
    return api.post<Report>(`/reports/schedules/${id}/trigger`);
  },

  async getScheduleHistory(id: string): Promise<Array<{
    runAt: string;
    status: string;
    reportId?: string;
    error?: string;
  }>> {
    return api.get(`/reports/schedules/${id}/history`);
  },

  // Statistics
  async getReportStats(): Promise<{
    total: number;
    completed: number;
    failed: number;
    pending: number;
    byType: Record<string, number>;
    byFormat: Record<string, number>;
    averageGenerationTime: number;
  }> {
    return api.get('/reports/stats');
  },

  async getGenerationTrend(days: number = 30): Promise<Array<{
    date: string;
    total: number;
    completed: number;
    failed: number;
  }>> {
    return api.get(`/reports/trend?days=${days}`);
  },

  async getStorageStats(): Promise<{
    totalSize: number;
    reportCount: number;
    byType: Record<string, { count: number; size: number }>;
    oldestReport: string;
    largestReport: { id: string; size: number };
  }> {
    return api.get('/reports/storage/stats');
  },

  // Report Types
  async getReportTypes(): Promise<Array<{
    id: string;
    name: string;
    description: string;
    category: string;
    availableFormats: string[];
    parametersSchema: Record<string, unknown>;
  }>> {
    return api.get('/reports/types');
  },

  // WORM Storage
  async getWormStorageInfo(): Promise<{
    enabled: boolean;
    retentionDays: number;
    totalReports: number;
    lockedReports: number;
    storageUsed: number;
  }> {
    return api.get('/reports/storage/worm-info');
  },

  async extendRetention(id: string, days: number): Promise<Report> {
    return api.post<Report>(`/reports/${id}/extend-retention`, { days });
  },

  // Audit Trail
  async getReportAuditTrail(reportId: string): Promise<Array<{
    action: string;
    performedBy: string;
    performedAt: string;
    details: string;
  }>> {
    return api.get(`/reports/${reportId}/audit-trail`);
  },

  // Quick Reports
  async generateQuickReport(type: string, startDate: string, endDate: string, format: string): Promise<Report> {
    return api.post<Report>('/reports/quick', {
      type,
      startDate,
      endDate,
      format,
    });
  },

  async getRecentReports(limit: number = 5): Promise<Report[]> {
    return api.get<Report[]>(`/reports/recent?limit=${limit}`);
  },

  async getFrequentReports(limit: number = 10): Promise<Array<{
    type: string;
    count: number;
    lastGenerated: string;
  }>> {
    return api.get(`/reports/frequent?limit=${limit}`);
  },
};

export default reportingService;
