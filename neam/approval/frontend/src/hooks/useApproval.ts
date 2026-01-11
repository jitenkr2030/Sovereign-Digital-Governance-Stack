// React Hook for Approval System
// Provides reactive state management for approval operations

import { useState, useCallback, useEffect } from 'react';
import approvalApi from '../services/approvalApi';
import {
  ApprovalRequest,
  ApprovalDefinition,
  Delegation,
  AuditTrailEntry,
  AuditReport,
  ApprovalSummary,
  ApprovalStatus,
  ActionType,
  FilterParams,
} from '../types/approval';

interface UseApprovalOptions {
  autoRefresh?: boolean;
  refreshInterval?: number;
}

interface UseApprovalReturn {
  // State
  requests: ApprovalRequest[];
  definitions: ApprovalDefinition[];
  delegations: Delegation[];
  pendingApprovals: ApprovalRequest[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  fetchRequests: (params?: FilterParams) => Promise<void>;
  fetchPendingApprovals: (userId: string) => Promise<void>;
  fetchDefinitions: (category?: string) => Promise<void>;
  fetchDelegations: (userId: string) => Promise<void>;
  createRequest: (request: any) => Promise<string>;
  submitDecision: (requestId: string, approverId: string, decision: ActionType, comments?: string) => Promise<void>;
  createDelegation: (delegation: any) => Promise<string>;
  revokeDelegation: (delegationId: string, reason: string) => Promise<void>;
  
  // Utility
  clearError: () => void;
  refresh: () => Promise<void>;
}

export function useApproval(options: UseApprovalOptions = {}): UseApprovalReturn {
  const { autoRefresh = false, refreshInterval = 30000 } = options;

  // State
  const [requests, setRequests] = useState<ApprovalRequest[]>([]);
  const [definitions, setDefinitions] = useState<ApprovalDefinition[]>([]);
  const [delegations, setDelegations] = useState<Delegation[]>([]);
  const [pendingApprovals, setPendingApprovals] = useState<ApprovalRequest[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Clear error
  const clearError = useCallback(() => {
    setError(null);
  }, []);

  // Fetch approval requests
  const fetchRequests = useCallback(async (params?: FilterParams) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await approvalApi.getApprovalRequests(params);
      setRequests(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch requests');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch pending approvals for user
  const fetchPendingApprovals = useCallback(async (userId: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await approvalApi.getPendingApprovals(userId);
      setPendingApprovals(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch pending approvals');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch approval definitions
  const fetchDefinitions = useCallback(async (category?: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await approvalApi.getDefinitions({ category });
      setDefinitions(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch definitions');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Fetch delegations
  const fetchDelegations = useCallback(async (userId: string) => {
    setIsLoading(true);
    setError(null);
    try {
      const data = await approvalApi.getDelegations(userId);
      setDelegations(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch delegations');
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Create approval request
  const createRequest = useCallback(async (request: any): Promise<string> => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await approvalApi.createApprovalRequest(request);
      return result.request_id;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create request';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Submit decision
  const submitDecision = useCallback(async (
    requestId: string,
    approverId: string,
    decision: ActionType,
    comments?: string
  ) => {
    setIsLoading(true);
    setError(null);
    try {
      await approvalApi.submitDecision(requestId, {
        approver_id: approverId,
        decision,
        comments,
      });
      // Refresh pending approvals after decision
      await fetchPendingApprovals(approverId);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to submit decision';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [fetchPendingApprovals]);

  // Create delegation
  const createDelegation = useCallback(async (delegation: any): Promise<string> => {
    setIsLoading(true);
    setError(null);
    try {
      const result = await approvalApi.createDelegation(delegation);
      return result.delegation_id;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create delegation';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Revoke delegation
  const revokeDelegation = useCallback(async (delegationId: string, reason: string) => {
    setIsLoading(true);
    setError(null);
    try {
      await approvalApi.revokeDelegation(delegationId, reason);
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to revoke delegation';
      setError(message);
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Refresh all data
  const refresh = useCallback(async () => {
    await Promise.all([
      fetchRequests(),
      fetchDefinitions(),
    ]);
  }, [fetchRequests, fetchDefinitions]);

  // Auto refresh
  useEffect(() => {
    if (autoRefresh) {
      const interval = setInterval(refresh, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [autoRefresh, refreshInterval, refresh]);

  return {
    requests,
    definitions,
    delegations,
    pendingApprovals,
    isLoading,
    error,
    fetchRequests,
    fetchPendingApprovals,
    fetchDefinitions,
    fetchDelegations,
    createRequest,
    submitDecision,
    createDelegation,
    revokeDelegation,
    clearError,
    refresh,
  };
}

// Hook for single approval request
export function useApprovalRequest(requestId: string | null) {
  const [request, setRequest] = useState<ApprovalRequest | null>(null);
  const [approvers, setApprovers] = useState<any[]>([]);
  const [actions, setActions] = useState<any[]>([]);
  const [summary, setSummary] = useState<ApprovalSummary | null>(null);
  const [auditTrail, setAuditTrail] = useState<AuditTrailEntry[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchRequest = useCallback(async () => {
    if (!requestId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await approvalApi.getApprovalRequest(requestId);
      setRequest(data.request);
      setApprovers(data.approvers);
      setActions(data.actions);
      
      const summaryData = await approvalApi.getApprovalSummary(requestId);
      setSummary(summaryData);
      
      const auditData = await approvalApi.getAuditTrail(requestId);
      setAuditTrail(auditData.audit_trail);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch request');
    } finally {
      setIsLoading(false);
    }
  }, [requestId]);

  useEffect(() => {
    fetchRequest();
  }, [fetchRequest]);

  return {
    request,
    approvers,
    actions,
    summary,
    auditTrail,
    isLoading,
    error,
    refresh: fetchRequest,
  };
}

// Hook for audit operations
export function useAudit(requestId: string | null) {
  const [auditTrail, setAuditTrail] = useState<AuditTrailEntry[]>([]);
  const [verification, setVerification] = useState<any>(null);
  const [report, setReport] = useState<AuditReport | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchAuditTrail = useCallback(async () => {
    if (!requestId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await approvalApi.getAuditTrail(requestId);
      setAuditTrail(data.audit_trail);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch audit trail');
    } finally {
      setIsLoading(false);
    }
  }, [requestId]);

  const verifyAuditTrail = useCallback(async () => {
    if (!requestId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await approvalApi.verifyAuditTrail(requestId);
      setVerification(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify audit trail');
    } finally {
      setIsLoading(false);
    }
  }, [requestId]);

  const generateReport = useCallback(async (generatedBy: string) => {
    if (!requestId) return;
    
    setIsLoading(true);
    setError(null);
    
    try {
      const data = await approvalApi.generateAuditReport(requestId, generatedBy);
      setReport(data);
      return data;
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to generate report');
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, [requestId]);

  const downloadReport = useCallback(async (format: 'pdf' | 'json' | 'csv') => {
    if (!requestId) return;
    
    try {
      const blob = await approvalApi.downloadAuditReport(requestId, format);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit-report-${requestId}.${format}`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to download report');
    }
  }, [requestId]);

  return {
    auditTrail,
    verification,
    report,
    isLoading,
    error,
    fetchAuditTrail,
    verifyAuditTrail,
    generateReport,
    downloadReport,
  };
}
