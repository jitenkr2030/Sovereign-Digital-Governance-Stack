// User and Authentication Types
export interface User {
  id: string;
  email: string;
  name: string;
  role: 'admin' | 'regulator' | 'auditor' | 'viewer';
  permissions: Permission[];
  avatar?: string;
  lastLogin?: string;
  createdAt: string;
}

export type Permission = 
  | 'licenses:read'
  | 'licenses:write'
  | 'licenses:delete'
  | 'energy:read'
  | 'energy:write'
  | 'reports:read'
  | 'reports:write'
  | 'reports:delete'
  | 'settings:read'
  | 'settings:write';

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
  expiresAt: string;
}

// License Types
export interface License {
  id: string;
  entityId: string;
  entityName: string;
  entityType: 'VASP' | 'CASP' | 'OTHER';
  licenseType: 'exchange' | 'custodian' | 'wallet' | 'mining' | 'other';
  licenseNumber: string;
  status: LicenseStatus;
  issueDate: string;
  expiryDate: string;
  jurisdiction: string;
  regulatoryBody: string;
  conditions: string[];
  restrictions: string[];
  complianceScore: number;
  lastAuditDate?: string;
  nextAuditDate?: string;
  documents: LicenseDocument[];
  history: LicenseHistoryEntry[];
  createdAt: string;
  updatedAt: string;
}

export type LicenseStatus = 
  | 'active'
  | 'pending'
  | 'suspended'
  | 'revoked'
  | 'expired'
  | 'under_review';

export interface LicenseDocument {
  id: string;
  name: string;
  type: string;
  url: string;
  uploadedAt: string;
}

export interface LicenseHistoryEntry {
  id: string;
  action: string;
  performedBy: string;
  performedAt: string;
  details: string;
}

export interface LicenseApplication {
  id: string;
  entityName: string;
  entityType: 'VASP' | 'CASP' | 'OTHER';
  licenseType: string;
  jurisdiction: string;
  status: 'submitted' | 'under_review' | 'approved' | 'rejected';
  submittedAt: string;
  documents: string[];
  notes: string;
}

// Energy Types
export interface EnergyMetrics {
  timestamp: string;
  region: string;
  consumption: number;
  demand: number;
  supply: number;
  renewablePercentage: number;
  carbonEmissions: number;
  gridFrequency: number;
  voltage: number;
}

export interface EnergyForecast {
  timestamp: string;
  region: string;
  predictedLoad: number;
  confidenceLower: number;
  confidenceUpper: number;
  actualLoad?: number;
}

export interface RegionalEnergyData {
  region: string;
  totalConsumption: number;
  peakDemand: number;
  averageLoad: number;
  renewablePercentage: number;
  numberOfNodes: number;
  carbonIntensity: number;
}

export interface GridStatus {
  region: string;
  status: 'normal' | 'warning' | 'critical';
  frequency: number;
  loadPercentage: number;
  availableCapacity: number;
  lastUpdated: string;
}

// Report Types
export interface Report {
  id: string;
  title: string;
  type: ReportType;
  format: 'pdf' | 'csv' | 'xlsx' | 'json';
  status: ReportStatus;
  parameters: ReportParameters;
  filePath?: string;
  fileSize?: number;
  generatedBy: string;
  entityId?: string;
  entityType?: string;
  createdAt: string;
  completedAt?: string;
  expiresAt?: string;
  error?: string;
}

export type ReportType = 
  | 'compliance'
  | 'financial'
  | 'energy'
  | 'audit'
  | 'regulatory'
  | 'custom';

export type ReportStatus = 
  | 'pending'
  | 'processing'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface ReportParameters {
  startDate?: string;
  endDate?: string;
  entityIds?: string[];
  reportTemplate?: string;
  includeCharts?: boolean;
  includeSummary?: boolean;
}

export interface ReportTemplate {
  id: string;
  name: string;
  description: string;
  type: ReportType;
  parametersSchema: Record<string, unknown>;
  defaultParameters: Record<string, unknown>;
  outputFormat: 'pdf' | 'csv' | 'xlsx';
  isActive: boolean;
}

export interface ReportSchedule {
  id: string;
  name: string;
  reportType: ReportType;
  frequency: 'daily' | 'weekly' | 'monthly' | 'quarterly' | 'annually';
  cronExpression?: string;
  parameters: ReportParameters;
  recipients: string[];
  isActive: boolean;
  lastRun?: string;
  nextRun: string;
}

// Dashboard Types
export interface DashboardMetrics {
  totalLicenses: number;
  activeLicenses: number;
  pendingApplications: number;
  complianceScore: number;
  totalEnergyConsumption: number;
  averageGridLoad: number;
  activeAlerts: number;
  reportsGenerated: number;
  recentActivity: ActivityItem[];
}

export interface ActivityItem {
  id: string;
  type: 'license' | 'energy' | 'report' | 'compliance' | 'system';
  action: string;
  description: string;
  entityId?: string;
  entityType?: string;
  performedBy?: string;
  timestamp: string;
}

export interface Alert {
  id: string;
  type: 'warning' | 'error' | 'info' | 'success';
  category: 'license' | 'energy' | 'compliance' | 'system';
  title: string;
  message: string;
  entityId?: string;
  entityType?: string;
  isRead: boolean;
  createdAt: string;
  acknowledgedAt?: string;
  acknowledgedBy?: string;
}

// API Response Types
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  meta?: PaginationMeta;
}

export interface PaginationMeta {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  meta: PaginationMeta;
}

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

// Filter and Sort Types
export interface FilterOptions {
  search?: string;
  status?: string[];
  type?: string[];
  dateRange?: {
    start: string;
    end: string;
  };
  jurisdiction?: string[];
}

export interface SortOptions {
  field: string;
  direction: 'asc' | 'desc';
}

// Notification Types
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  duration?: number;
  action?: {
    label: string;
    onClick: () => void;
  };
}

// Theme Types
export interface ThemeSettings {
  mode: 'light' | 'dark' | 'system';
  primaryColor: string;
  compactMode: boolean;
}

// Chart Data Types
export interface TimeSeriesDataPoint {
  timestamp: string;
  value: number;
  label?: string;
}

export interface ChartDataPoint {
  x: number | string | Date;
  y: number;
  label?: string;
  color?: string;
}

export interface PieChartData {
  name: string;
  value: number;
  color?: string;
  percentage?: number;
}
