// CSIC Platform - Frontend API Client
// Centralized API communication layer for the regulator dashboard

import axios, { AxiosInstance, AxiosError } from 'axios';

// API Configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';
const API_TIMEOUT = 30000;

// Custom error class for API errors
export class APIError extends Error {
  constructor(
    message: string,
    public statusCode: number,
    public code: string,
    public details?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'APIError';
  }
}

// Response types
export interface ApiResponse<T> {
  data: T;
  meta?: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
  timestamp: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
}

// Generate unique request ID
const generateRequestId = (): string => {
  return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
};

// Create axios instance with default configuration
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: API_BASE_URL,
    timeout: API_TIMEOUT,
    headers: {
      'Content-Type': 'application/json',
      'X-Request-ID': generateRequestId(),
    },
  });

  // Request interceptor - Add auth token
  client.interceptors.request.use(
    (config) => {
      const token = localStorage.getItem('auth_token');
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      config.headers['X-Request-ID'] = generateRequestId();
      return config;
    },
    (error) => Promise.reject(error)
  );

  // Response interceptor - Handle errors and responses
  client.interceptors.response.use(
    (response) => response,
    (error: AxiosError) => {
      if (error.response) {
        const { status } = error.response;
        
        // Handle specific error codes
        if (status === 401) {
          localStorage.removeItem('auth_token');
          localStorage.removeItem('user_data');
          window.location.href = '/login';
        }
        
        const data = error.response.data as { message?: string; code?: string; details?: Record<string, unknown> };
        throw new APIError(
          data?.message || 'An error occurred',
          status || 0,
          data?.code || 'UNKNOWN_ERROR',
          data?.details
        );
      } else if (error.request) {
        throw new APIError('Network error - please check your connection', 0, 'NETWORK_ERROR');
      } else {
        throw new APIError(error.message || 'An unexpected error occurred', 0, 'UNKNOWN_ERROR');
      }
    }
  );

  return client;
};

// API Client instance
export const apiClient = createApiClient();

// Security Service API
export const SecurityAPI = {
  // Audit Logs
  async getAuditLogs(params?: {
    page?: number;
    pageSize?: number;
    startDate?: string;
    endDate?: string;
    eventType?: string;
    severity?: string;
    userId?: string;
    entityId?: string;
    search?: string;
  }) {
    const response = await apiClient.get('/security/audit/logs', { params });
    return response.data as PaginatedResponse<{
      id: string;
      timestamp: string;
      eventType: string;
      severity: string;
      entityType: string;
      entityId: string;
      userId: string;
      userEmail: string;
      action: string;
      prevHash: string;
      currentHash: string;
    }>;
  },

  async getAuditLogById(id: string) {
    const response = await apiClient.get(`/security/audit/logs/${id}`);
    return response.data;
  },

  async verifyChain(startId: string, endId: string) {
    const response = await apiClient.post('/security/audit/verify', { startId, endId });
    return response.data;
  },

  // Key Management
  async getKeys() {
    const response = await apiClient.get('/security/keys');
    return response.data;
  },

  async getKeyById(id: string) {
    const response = await apiClient.get(`/security/keys/${id}`);
    return response.data;
  },

  async rotateKey(id: string) {
    const response = await apiClient.post(`/security/keys/${id}/rotate`);
    return response.data;
  },

  async revokeKey(id: string, reason: string) {
    const response = await apiClient.post(`/security/keys/${id}/revoke`, { reason });
    return response.data;
  },

  // SIEM Events
  async getSIEMEvents(params?: {
    page?: number;
    pageSize?: number;
    severity?: string;
    status?: string;
    eventType?: string;
  }) {
    const response = await apiClient.get('/security/siem/events', { params });
    return response.data as PaginatedResponse<{
      id: string;
      eventType: string;
      severity: string;
      source: string;
      sourceIp: string;
      action: string;
      outcome: string;
      timestamp: string;
      status: string;
    }>;
  },

  async getSIEMStats() {
    const response = await apiClient.get('/security/siem/stats');
    return response.data;
  },

  async retrySIEMEvent(id: string) {
    const response = await apiClient.post(`/security/siem/events/${id}/retry`);
    return response.data;
  },

  // Compliance
  async getComplianceScore() {
    const response = await apiClient.get('/security/compliance/score');
    return response.data;
  },

  async getComplianceCategories() {
    const response = await apiClient.get('/security/compliance/categories');
    return response.data;
  },

  // Licenses
  async getLicenses() {
    const response = await apiClient.get('/security/licenses');
    return response.data;
  },

  async getLicenseById(id: string) {
    const response = await apiClient.get(`/security/licenses/${id}`);
    return response.data;
  },

  // Dashboard
  async getDashboardStats() {
    const response = await apiClient.get('/security/dashboard/stats');
    return response.data;
  },

  async getComplianceTrend(period: string) {
    const response = await apiClient.get('/security/dashboard/compliance-trend', { params: { period } });
    return response.data;
  },

  async getEventTimeline(params?: { hours?: number; limit?: number }) {
    const response = await apiClient.get('/security/dashboard/event-timeline', { params });
    return response.data;
  },
};

// Dashboard API
export const DashboardAPI = {
  async getSystemStatus() {
    const response = await apiClient.get('/dashboard/system-status');
    return response.data;
  },

  async getMetrics() {
    const response = await apiClient.get('/dashboard/metrics');
    return response.data;
  },

  async getCharts(type: string, period: string) {
    const response = await apiClient.get(`/dashboard/charts/${type}`, {
      params: { period },
    });
    return response.data;
  },
};

// Alert API
export const AlertAPI = {
  async getAlerts(params?: {
    page?: number;
    pageSize?: number;
    severity?: string;
    status?: string;
    category?: string;
    startDate?: string;
    endDate?: string;
  }) {
    const response = await apiClient.get('/alerts', { params });
    return response.data as PaginatedResponse<{
      id: string;
      title: string;
      description: string;
      severity: string;
      status: string;
      category: string;
      source: string;
      createdAt: string;
      updatedAt: string;
    }>;
  },

  async getAlertById(id: string) {
    const response = await apiClient.get(`/alerts/${id}`);
    return response.data;
  },

  async acknowledgeAlert(id: string) {
    const response = await apiClient.post(`/alerts/${id}/acknowledge`);
    return response.data;
  },

  async resolveAlert(id: string, resolution: string) {
    const response = await apiClient.post(`/alerts/${id}/resolve`, { resolution });
    return response.data;
  },

  async getAlertStats() {
    const response = await apiClient.get('/alerts/stats');
    return response.data;
  },
};

// Compliance API
export const ComplianceAPI = {
  async getComplianceReports(params?: {
    page?: number;
    pageSize?: number;
    entityType?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/compliance/reports', { params });
    return response.data;
  },

  async getComplianceReportById(id: string) {
    const response = await apiClient.get(`/compliance/reports/${id}`);
    return response.data;
  },

  async generateComplianceReport(entityType: string, entityId: string, period: string) {
    const response = await apiClient.post('/compliance/reports/generate', {
      entityType,
      entityId,
      period,
    });
    return response.data;
  },

  async getViolations(params?: {
    page?: number;
    pageSize?: number;
    severity?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/compliance/violations', { params });
    return response.data;
  },
};

// Reporting API
export const ReportingAPI = {
  async getReports(params?: {
    page?: number;
    pageSize?: number;
    type?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/reports', { params });
    return response.data;
  },

  async getReportById(id: string) {
    const response = await apiClient.get(`/reports/${id}`);
    return response.data;
  },

  async generateReport(data: {
    type: string;
    title: string;
    parameters: Record<string, unknown>;
  }) {
    const response = await apiClient.post('/reports/generate', data);
    return response.data;
  },

  async exportReport(id: string, format: 'pdf' | 'xlsx' | 'csv') {
    const response = await apiClient.get(`/reports/${id}/export`, {
      params: { format },
      responseType: 'blob',
    });
    return response.data;
  },
};

// Audit API
export const AuditAPI = {
  async getAuditLogs(params?: {
    page?: number;
    pageSize?: number;
    userId?: string;
    action?: string;
    resourceType?: string;
    startDate?: string;
    endDate?: string;
  }) {
    const response = await apiClient.get('/audit/logs', { params });
    return response.data;
  },

  async getAuditLogById(id: string) {
    const response = await apiClient.get(`/audit/logs/${id}`);
    return response.data;
  },

  async exportAuditLogs(params: {
    startDate: string;
    endDate: string;
    format: 'pdf' | 'xlsx' | 'csv';
  }) {
    const response = await apiClient.get('/audit/export', {
      params,
      responseType: 'blob',
    });
    return response.data;
  },
};

// User API
export const UserAPI = {
  async getCurrentUser() {
    const response = await apiClient.get('/users/me');
    return response.data;
  },

  async updateProfile(data: {
    name?: string;
    email?: string;
    phone?: string;
  }) {
    const response = await apiClient.patch('/users/me/profile', data);
    return response.data;
  },

  async getUsers(params?: {
    page?: number;
    pageSize?: number;
    role?: string;
    status?: string;
  }) {
    const response = await apiClient.get('/users', { params });
    return response.data;
  },

  async createUser(data: {
    username: string;
    email: string;
    role: string;
    permissions?: string[];
  }) {
    const response = await apiClient.post('/users', data);
    return response.data;
  },
};

// Health API
export const HealthAPI = {
  async getSystemHealth() {
    const response = await apiClient.get('/health');
    return response.data;
  },

  async getComponentHealth(component: string) {
    const response = await apiClient.get(`/health/${component}`);
    return response.data;
  },
};

// Export all API modules
export const api = {
  security: SecurityAPI,
  dashboard: DashboardAPI,
  alerts: AlertAPI,
  compliance: ComplianceAPI,
  reports: ReportingAPI,
  audit: AuditAPI,
  users: UserAPI,
  health: HealthAPI,
};

export default api;
