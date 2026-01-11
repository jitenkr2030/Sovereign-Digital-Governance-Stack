// User and Authentication Types
export type UserRole = 'admin' | 'regulator' | 'auditor' | 'user';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  organization: string;
  avatar?: string;
  permissions: string[];
  lastLogin: string;
  mfaEnabled: boolean;
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  token: string | null;
}

export interface LoginCredentials {
  email: string;
  password: string;
  mfaCode?: string;
}

export interface AuthResponse {
  user: User;
  token: string;
  refreshToken: string;
  expiresAt: string;
}

// Audit Log Types
export interface AuditLog {
  id: string;
  timestamp: string;
  eventType: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  entityType: string;
  entityId: string;
  userId: string;
  userEmail: string;
  action: string;
  resourceType: string;
  resourceId: string;
  oldValue?: string;
  newValue?: string;
  metadata?: string;
  ipAddress: string;
  userAgent: string;
  correlationId: string;
  sessionId: string;
  prevHash: string;
  currentHash: string;
  createdAt: string;
}

export interface AuditLogFilters {
  startDate?: string;
  endDate?: string;
  eventType?: string;
  severity?: string;
  userId?: string;
  entityId?: string;
  searchQuery?: string;
}

// Compliance Types
export interface ComplianceScore {
  overall: number;
  categories: ComplianceCategory[];
  trend: 'up' | 'down' | 'stable';
  lastUpdated: string;
}

export interface ComplianceCategory {
  name: string;
  score: number;
  weight: number;
  issues: number;
}

export interface License {
  id: string;
  type: string;
  status: 'active' | 'expired' | 'revoked' | 'pending';
  issuedAt: string;
  expiresAt: string;
  holder: string;
  jurisdiction: string;
  conditions: string[];
}

// Key Management Types
export interface CryptoKey {
  id: string;
  name: string;
  purpose: 'encryption' | 'signing' | 'verification' | 'audit';
  algorithm: string;
  status: 'active' | 'revoked' | 'expired';
  createdAt: string;
  lastUsed?: string;
  expiresAt?: string;
  publicKey?: string;
}

export interface KeyRotation {
  keyId: string;
  oldVersion: number;
  newVersion: number;
  rotatedAt: string;
  rotatedBy: string;
}

// SIEM Types
export interface SIEMEvent {
  id: string;
  eventType: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  source: string;
  sourceIp?: string;
  destinationIp?: string;
  userId?: string;
  sessionId?: string;
  action: string;
  outcome: 'success' | 'failure' | 'blocked';
  timestamp: string;
  status: 'pending' | 'forwarded' | 'failed' | 'retrying';
  eventData: Record<string, unknown>;
}

export interface SIEMStats {
  totalEvents: number;
  pendingEvents: number;
  forwardedEvents: number;
  failedEvents: number;
  retryingEvents: number;
  oldestEvent?: string;
  newestEvent?: string;
}

// Chain Verification Types
export interface ChainVerificationResult {
  isValid: boolean;
  startId: string;
  endId: string;
  entriesVerified: number;
  verifiedAt: string;
  errors?: string[];
}

// Dashboard Types
export interface DashboardWidget {
  id: string;
  type: 'compliance' | 'licenses' | 'alerts' | 'keys' | 'chart' | 'table';
  title: string;
  size: 'small' | 'medium' | 'large';
  position: { x: number; y: number };
  config?: Record<string, unknown>;
}

export interface DashboardStats {
  totalUsers: number;
  activeLicenses: number;
  complianceScore: number;
  pendingAlerts: number;
  keysExpiringSoon: number;
  recentEvents: number;
}

// API Response Types
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  pagination?: PaginationInfo;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

// Notification Types
export interface Notification {
  id: string;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  message: string;
  read: boolean;
  createdAt: string;
  actionUrl?: string;
}

// Report Types
export interface ComplianceReport {
  id: string;
  title: string;
  period: { start: string; end: string };
  generatedAt: string;
  generatedBy: string;
  status: 'draft' | 'final' | 'archived';
  sections: ReportSection[];
}

export interface ReportSection {
  name: string;
  type: 'summary' | 'details' | 'charts' | 'recommendations';
  content: Record<string, unknown>;
}
