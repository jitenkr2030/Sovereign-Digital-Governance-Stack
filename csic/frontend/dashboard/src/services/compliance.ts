import apiClient from './api';

// Types
export interface ComplianceStatus {
  overallScore: number;
  status: 'compliant' | 'partial' | 'non-compliant' | 'under-review';
  lastChecked: string;
  nextReviewDate: string;
  categories: ComplianceCategory[];
}

export interface ComplianceCategory {
  id: string;
  name: string;
  score: number;
  status: 'compliant' | 'partial' | 'non-compliant' | 'under-review';
  requirementsCount: number;
  compliantCount: number;
}

export interface ComplianceChecklist {
  id: string;
  categoryId: string;
  title: string;
  description: string;
  status: 'compliant' | 'partial' | 'non-compliant' | 'pending' | 'not-applicable';
  priority: 'high' | 'medium' | 'low';
  lastChecked: string;
  checkedBy?: string;
  notes?: string;
  evidence?: EvidenceItem[];
}

export interface EvidenceItem {
  id: string;
  name: string;
  type: 'document' | 'image' | 'certificate' | 'report';
  uploadedAt: string;
  uploadedBy: string;
  url: string;
}

export interface Violation {
  id: string;
  title: string;
  description: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  category: string;
  status: 'open' | 'in-progress' | 'resolved' | 'waived';
  detectedAt: string;
  dueDate: string;
  assignedTo?: string;
  affectedEntity: {
    type: string;
    id: string;
    name: string;
  };
}

export interface ExemptionRequest {
  id: string;
  requirementId: string;
  requirementTitle: string;
  reason: string;
  supportingDocuments: string[];
  requestedBy: string;
  requestedAt: string;
  status: 'pending' | 'approved' | 'rejected';
  reviewedBy?: string;
  reviewedAt?: string;
  reviewNotes?: string;
  expiryDate?: string;
}

export interface ExemptionPayload {
  requirementId: string;
  reason: string;
  supportingDocuments: string[];
  requestedDuration?: string;
}

export interface ComplianceReport {
  id: string;
  title: string;
  type: 'full' | 'summary' | 'category' | 'violation';
  generatedAt: string;
  generatedBy: string;
  period: {
    start: string;
    end: string;
  };
  status: 'compliant' | 'partial' | 'non-compliant';
  downloadUrl: string;
}

export interface AuditTrailEntry {
  id: string;
  action: string;
  entityType: string;
  entityId: string;
  entityName: string;
  performedBy: string;
  performedAt: string;
  details: string;
  previousValue?: string;
  newValue?: string;
  ipAddress?: string;
}

export interface AuditTrailFilter {
  entityType?: string;
  entityId?: string;
  performedBy?: string;
  startDate?: string;
  endDate?: string;
  action?: string;
  page?: number;
  pageSize?: number;
}

// API Service
export const ComplianceService = {
  // Get overall compliance status
  getStatus: async (): Promise<ComplianceStatus> => {
    const response = await apiClient.get('/compliance/status');
    return response.data;
  },

  // Get compliance checklist
  getChecklist: async (categoryId?: string): Promise<ComplianceChecklist[]> => {
    const params = categoryId ? { categoryId } : {};
    const response = await apiClient.get('/compliance/checklist', { params });
    return response.data;
  },

  // Update checklist item status
  updateChecklistItem: async (
    itemId: string,
    status: ComplianceChecklist['status'],
    notes?: string
  ): Promise<ComplianceChecklist> => {
    const response = await apiClient.patch(`/compliance/checklist/${itemId}`, {
      status,
      notes,
    });
    return response.data;
  },

  // Submit compliance check
  submitCheck: async (
    checklistIds: string[]
  ): Promise<{ checkId: string; status: string }> => {
    const response = await apiClient.post('/compliance/check', { checklistIds });
    return response.data;
  },

  // Get violations
  getViolations: async (
    filters?: Partial<{
      status: Violation['status'];
      severity: Violation['severity'];
      category: string;
      page: number;
      pageSize: number;
    }>
  ): Promise<{
    items: Violation[];
    total: number;
    page: number;
    pageSize: number;
  }> => {
    const response = await apiClient.get('/compliance/violations', { params: filters });
    return response.data;
  },

  // Create violation
  createViolation: async (
    violation: Omit<Violation, 'id' | 'detectedAt'>
  ): Promise<Violation> => {
    const response = await apiClient.post('/compliance/violations', violation);
    return response.data;
  },

  // Update violation
  updateViolation: async (
    violationId: string,
    updates: Partial<Violation>
  ): Promise<Violation> => {
    const response = await apiClient.patch(
      `/compliance/violations/${violationId}`,
      updates
    );
    return response.data;
  },

  // Get exemption requests
  getExemptions: async (
    status?: ExemptionRequest['status']
  ): Promise<ExemptionRequest[]> => {
    const params = status ? { status } : {};
    const response = await apiClient.get('/compliance/exemptions', { params });
    return response.data;
  },

  // Request exemption
  requestExemption: async (
    payload: ExemptionPayload
  ): Promise<ExemptionRequest> => {
    const response = await apiClient.post('/compliance/exemption', payload);
    return response.data;
  },

  // Approve exemption
  approveExemption: async (
    exemptionId: string,
    notes: string,
    expiryDate?: string
  ): Promise<ExemptionRequest> => {
    const response = await apiClient.post(
      `/compliance/exemption/${exemptionId}/approve`,
      { notes, expiryDate }
    );
    return response.data;
  },

  // Reject exemption
  rejectExemption: async (
    exemptionId: string,
    reason: string
  ): Promise<ExemptionRequest> => {
    const response = await apiClient.post(
      `/compliance/exemption/${exemptionId}/reject`,
      { reason }
    );
    return response.data;
  },

  // Get compliance reports
  getReports: async (
    filters?: Partial<{
      type: ComplianceReport['type'];
      startDate: string;
      endDate: string;
      page: number;
      pageSize: number;
    }>
  ): Promise<{
    items: ComplianceReport[];
    total: number;
  }> => {
    const response = await apiClient.get('/compliance/reports', { params: filters });
    return response.data;
  },

  // Generate compliance report
  generateReport: async (
    reportType: ComplianceReport['type'],
    period: { start: string; end: string },
    categoryId?: string
  ): Promise<ComplianceReport> => {
    const response = await apiClient.post('/compliance/reports/generate', {
      type: reportType,
      period,
      categoryId,
    });
    return response.data;
  },

  // Download report
  downloadReport: async (reportId: string): Promise<Blob> => {
    const response = await apiClient.get(`/compliance/reports/${reportId}/download`, {
      responseType: 'blob',
    });
    return response.data;
  },

  // Get audit trail
  getAuditTrail: async (
    filter: AuditTrailFilter
  ): Promise<{
    items: AuditTrailEntry[];
    total: number;
  }> => {
    const response = await apiClient.get('/compliance/audit-trail', {
      params: filter,
    });
    return response.data;
  },

  // Export audit trail
  exportAuditTrail: async (
    filter: AuditTrailFilter,
    format: 'csv' | 'pdf' | 'json'
  ): Promise<Blob> => {
    const response = await apiClient.get('/compliance/audit-trail/export', {
      params: { ...filter, format },
      responseType: 'blob',
    });
    return response.data;
  },

  // Get compliance metrics
  getMetrics: async (
    timeRange: 'week' | 'month' | 'quarter' | 'year'
  ): Promise<{
    scoreTrend: { date: string; score: number }[];
    violationTrend: { date: string; count: number }[];
    categoryBreakdown: { category: string; score: number }[];
    topIssues: { issue: string; count: number }[];
  }> => {
    const response = await apiClient.get('/compliance/metrics', {
      params: { timeRange },
    });
    return response.data;
  },

  // Get upcoming reviews
  getUpcomingReviews: async (days: number = 30): Promise<{
    items: { id: string; title: string; dueDate: string; assignedTo?: string }[];
  }> => {
    const response = await apiClient.get('/compliance/upcoming-reviews', {
      params: { days },
    });
    return response.data;
  },

  // Bulk update checklist items
  bulkUpdateChecklist: async (
    updates: { itemId: string; status: ComplianceChecklist['status'] }[]
  ): Promise<{ updated: number; failed: number }> => {
    const response = await apiClient.post('/compliance/checklist/bulk-update', {
      updates,
    });
    return response.data;
  },
};

export default ComplianceService;
