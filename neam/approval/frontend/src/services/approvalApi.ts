// Approval System API Service
// Handles all API communication for the approval system

import {
  ApprovalRequest,
  ApprovalDefinition,
  ApprovalAction,
  ApproverAssignment,
  AuditTrailEntry,
  AuditReport,
  Delegation,
  CreateDefinitionRequest,
  CreateRequestRequest,
  DecisionRequest,
  CreateDelegationRequest,
  FilterParams,
  ApprovalSummary,
  ApiError,
} from '../types/approval';

const API_BASE = '/api/v1';

class ApprovalApiService {
  private getHeaders(): HeadersInit {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };
    const token = localStorage.getItem('auth_token');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    return headers;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const error: ApiError = await response.json().catch(() => ({
        code: 'UNKNOWN_ERROR',
        message: 'An unexpected error occurred',
      }));
      throw new Error(error.message);
    }
    return response.json();
  }

  // ============================================
  // Definition Endpoints
  // ============================================

  async createDefinition(request: CreateDefinitionRequest): Promise<{ id: string }> {
    const response = await fetch(`${API_BASE}/approval-definitions`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });
    return this.handleResponse(response);
  }

  async getDefinitions(params?: { category?: string; active?: boolean }): Promise<ApprovalDefinition[]> {
    const searchParams = new URLSearchParams();
    if (params?.category) searchParams.set('category', params.category);
    if (params?.active) searchParams.set('active', 'true');

    const query = searchParams.toString();
    const url = `${API_BASE}/approval-definitions${query ? `?${query}` : ''}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    const result = await this.handleResponse<{ definitions: ApprovalDefinition[] }>(response);
    return result.definitions;
  }

  async getDefinitionById(id: string): Promise<ApprovalDefinition> {
    const response = await fetch(`${API_BASE}/approval-definitions/${id}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async updateDefinition(id: string, request: Partial<CreateDefinitionRequest>): Promise<void> {
    const response = await fetch(`${API_BASE}/approval-definitions/${id}`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });
    await this.handleResponse(response);
  }

  async deactivateDefinition(id: string): Promise<void> {
    const response = await fetch(`${API_BASE}/approval-definitions/${id}/deactivate`, {
      method: 'POST',
      headers: this.getHeaders(),
    });
    await this.handleResponse(response);
  }

  // ============================================
  // Request Endpoints
  // ============================================

  async createApprovalRequest(request: CreateRequestRequest): Promise<{
    request_id: string;
    workflow_id: string;
    status: string;
    deadline: string;
  }> {
    const response = await fetch(`${API_BASE}/approval-requests`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });
    return this.handleResponse(response);
  }

  async getApprovalRequest(id: string): Promise<{
    request: ApprovalRequest;
    approvers: ApproverAssignment[];
    actions: ApprovalAction[];
  }> {
    const response = await fetch(`${API_BASE}/approval-requests/${id}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async getApprovalRequests(params?: FilterParams): Promise<ApprovalRequest[]> {
    const searchParams = new URLSearchParams();
    if (params?.page) searchParams.set('page', params.page.toString());
    if (params?.limit) searchParams.set('limit', params.limit.toString());
    if (params?.status) searchParams.set('status', params.status.join(','));

    const query = searchParams.toString();
    const url = `${API_BASE}/approval-requests${query ? `?${query}` : ''}`;

    const response = await fetch(url, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    const result = await this.handleResponse<{ requests: ApprovalRequest[] }>(response);
    return result.requests;
  }

  async getPendingApprovals(userId: string): Promise<ApprovalRequest[]> {
    const response = await fetch(`${API_BASE}/approvals/pending?user_id=${userId}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    const result = await this.handleResponse<{ pending_approvals: ApprovalRequest[] }>(response);
    return result.pending_approvals;
  }

  async getApprovalSummary(requestId: string): Promise<ApprovalSummary> {
    const response = await fetch(`${API_BASE}/approval-requests/${requestId}/summary`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async cancelApprovalRequest(id: string, reason: string): Promise<void> {
    const response = await fetch(`${API_BASE}/approval-requests/${id}/cancel`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ reason }),
    });
    await this.handleResponse(response);
  }

  // ============================================
  // Decision Endpoints
  // ============================================

  async submitDecision(requestId: string, decision: DecisionRequest): Promise<{
    request_id: string;
    decision: string;
    message: string;
  }> {
    const response = await fetch(`${API_BASE}/approval-requests/${requestId}/decide`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(decision),
    });
    return this.handleResponse(response);
  }

  async approve(requestId: string, approverId: string, comments?: string): Promise<void> {
    await this.submitDecision(requestId, {
      approver_id: approverId,
      decision: 'APPROVE',
      comments,
    });
  }

  async reject(requestId: string, approverId: string, comments: string): Promise<void> {
    await this.submitDecision(requestId, {
      approver_id: approverId,
      decision: 'REJECT',
      comments,
    });
  }

  async abstain(requestId: string, approverId: string, comments?: string): Promise<void> {
    await this.submitDecision(requestId, {
      approver_id: approverId,
      decision: 'ABSTAIN',
      comments,
    });
  }

  async delegateDecision(
    requestId: string,
    fromUserId: string,
    toUserId: string,
    comments?: string
  ): Promise<void> {
    await this.submitDecision(requestId, {
      approver_id: toUserId,
      decision: 'APPROVE',
      comments,
      delegated_by: fromUserId,
    });
  }

  // ============================================
  // Audit Endpoints
  // ============================================

  async getAuditTrail(requestId: string): Promise<{
    request_id: string;
    audit_trail: AuditTrailEntry[];
    hash_chain_valid: boolean;
  }> {
    const response = await fetch(`${API_BASE}/approval-requests/${requestId}/audit-trail`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async verifyAuditTrail(requestId: string): Promise<{
    request_id: string;
    actions_count: number;
    hash_chain_valid: boolean;
    broken_at_action?: string;
    verification_message: string;
  }> {
    const response = await fetch(`${API_BASE}/approval-requests/${requestId}/verify`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async generateAuditReport(
    requestId: string,
    generatedBy: string
  ): Promise<AuditReport> {
    const response = await fetch(`${API_BASE}/approval-requests/${requestId}/audit-report`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ generated_by: generatedBy }),
    });
    return this.handleResponse(response);
  }

  async downloadAuditReport(requestId: string, format: 'pdf' | 'json' | 'csv'): Promise<Blob> {
    const response = await fetch(
      `${API_BASE}/approval-requests/${requestId}/audit-report?format=${format}`,
      {
        method: 'GET',
        headers: this.getHeaders(),
      }
    );

    if (!response.ok) {
      throw new Error('Failed to download audit report');
    }

    return response.blob();
  }

  // ============================================
  // Delegation Endpoints
  // ============================================

  async createDelegation(request: CreateDelegationRequest): Promise<{ delegation_id: string }> {
    const response = await fetch(`${API_BASE}/delegations`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify(request),
    });
    return this.handleResponse(response);
  }

  async getDelegations(userId: string, asDelegator?: boolean): Promise<Delegation[]> {
    const params = new URLSearchParams({ user_id: userId });
    if (asDelegator) params.set('as_delegator', 'true');

    const response = await fetch(`${API_BASE}/delegations?${params}`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    const result = await this.handleResponse<{ delegations: Delegation[] }>(response);
    return result.delegations;
  }

  async revokeDelegation(delegationId: string, reason: string): Promise<void> {
    const response = await fetch(`${API_BASE}/delegations/${delegationId}/revoke`, {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ reason }),
    });
    await this.handleResponse(response);
  }

  async getActiveDelegationsForUser(userId: string): Promise<Delegation[]> {
    return this.getDelegations(userId);
  }

  async getDelegationsCreatedByUser(userId: string): Promise<Delegation[]> {
    return this.getDelegations(userId, true);
  }

  // ============================================
  // Notification Endpoints
  // ============================================

  async getNotificationPreferences(userId: string): Promise<{
    channels: string[];
    enabled_types: string[];
  }> {
    const response = await fetch(`${API_BASE}/users/${userId}/notification-preferences`, {
      method: 'GET',
      headers: this.getHeaders(),
    });
    return this.handleResponse(response);
  }

  async updateNotificationPreferences(
    userId: string,
    preferences: { channels: string[]; enabled_types: string[] }
  ): Promise<void> {
    const response = await fetch(`${API_BASE}/users/${userId}/notification-preferences`, {
      method: 'PUT',
      headers: this.getHeaders(),
      body: JSON.stringify(preferences),
    });
    await this.handleResponse(response);
  }
}

export const approvalApi = new ApprovalApiService();
export default approvalApi;
