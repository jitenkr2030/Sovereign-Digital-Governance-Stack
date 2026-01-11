// Incident Types
export type IncidentSeverity = 'critical' | 'high' | 'medium' | 'low' | 'info';
export type IncidentStatus = 'open' | 'investigating' | 'containment' | 'resolved' | 'closed';
export type IncidentType = 'security' | 'compliance' | 'technical' | 'operational' | 'fraud';
export type IncidentPriority = 1 | 2 | 3 | 4 | 5;

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: IncidentSeverity;
  status: IncidentStatus;
  type: IncidentType;
  priority: IncidentPriority;
  assignee?: TeamMember;
  reporter: TeamMember;
  affectedSystems: string[];
  indicators: IndicatorOfCompromise[];
  timeline: TimelineEvent[];
  relatedIncidents: string[];
  tags: string[];
  createdAt: Date;
  updatedAt: Date;
  resolvedAt?: Date;
  closedAt?: Date;
  dueDate?: Date;
  metrics: IncidentMetrics;
}

export interface TeamMember {
  id: string;
  name: string;
  email: string;
  role: string;
  avatar?: string;
  department: string;
}

export interface IndicatorOfCompromise {
  id: string;
  type: 'ip' | 'domain' | 'hash' | 'email' | 'wallet' | 'url' | 'file_path' | 'registry' | 'mutex';
  value: string;
  confidence: number;
  source: string;
  firstSeen: Date;
  lastSeen: Date;
  tags: string[];
}

export interface TimelineEvent {
  id: string;
  timestamp: Date;
  type: 'creation' | 'update' | 'comment' | 'evidence' | 'action' | 'escalation' | 'resolution';
  actor: TeamMember;
  description: string;
  metadata?: Record<string, unknown>;
  attachments?: Attachment[];
}

export interface Attachment {
  id: string;
  name: string;
  type: string;
  size: number;
  url: string;
  uploadedAt: Date;
  uploadedBy: TeamMember;
}

export interface IncidentMetrics {
  timeToDetect: number; // in minutes
  timeToRespond: number; // in minutes
  timeToResolve: number; // in minutes
  affectedUsers: number;
  estimatedImpact: number;
  recoveryCost: number;
}

// Filter and Query Types
export interface IncidentFilter {
  status?: IncidentStatus[];
  severity?: IncidentSeverity[];
  type?: IncidentType[];
  assignee?: string;
  dateRange?: {
    start: Date;
    end: Date;
  };
  search?: string;
  tags?: string[];
}

export interface IncidentStats {
  total: number;
  bySeverity: Record<IncidentSeverity, number>;
  byStatus: Record<IncidentStatus, number>;
  byType: Record<IncidentType, number>;
  averageResolutionTime: number;
  openIncidentsCount: number;
  criticalIncidentsCount: number;
}

// Dashboard Types
export interface DashboardData {
  stats: IncidentStats;
  recentIncidents: Incident[];
  trendData: TrendDataPoint[];
  topAffectedSystems: SystemImpact[];
  teamWorkload: TeamWorkload[];
}

export interface TrendDataPoint {
  date: string;
  critical: number;
  high: number;
  medium: number;
  low: number;
  resolved: number;
}

export interface SystemImpact {
  system: string;
  incidentCount: number;
  severity: IncidentSeverity;
  lastIncident: Date;
}

export interface TeamWorkload {
  member: TeamMember;
  assignedCount: number;
  resolvedCount: number;
  averageResolutionTime: number;
}

// API Response Types
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  errors?: string[];
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// Notification Types
export interface Notification {
  id: string;
  type: 'incident_created' | 'incident_updated' | 'incident_escalated' | 'comment_added' | 'evidence_added';
  incidentId: string;
  incidentTitle: string;
  message: string;
  read: boolean;
  createdAt: Date;
}
