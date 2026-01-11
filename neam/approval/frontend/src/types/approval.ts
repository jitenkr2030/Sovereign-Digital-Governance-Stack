// Multi-Level Approval System TypeScript Types
// This file defines all TypeScript interfaces for the approval system

import { UUID } from 'crypto';

// ============================================
// Enums and Constants
// ============================================

export type ApprovalStatus = 
  | 'PENDING' 
  | 'APPROVED' 
  | 'REJECTED' 
  | 'EXPIRED' 
  | 'CANCELLED';

export type ActionType = 
  | 'APPROVE' 
  | 'REJECT' 
  | 'ABSTAIN';

export type NotificationChannel = 
  | 'email' 
  | 'slack' 
  | 'in_app' 
  | 'sms';

export type Priority = 
  | 'low' 
  | 'normal' 
  | 'high' 
  | 'critical';

export const ApprovalStatusColors: Record<ApprovalStatus, string> = {
  PENDING: '#F59E0B',   // Amber
  APPROVED: '#10B981',  // Emerald
  REJECTED: '#EF4444',  // Red
  EXPIRED: '#6B7280',   // Gray
  CANCELLED: '#6B7280', // Gray
};

export const PriorityWeights: Record<Priority, number> = {
  low: 1,
  normal: 2,
  high: 3,
  critical: 4,
};

// ============================================
// Policy and Configuration Types
// ============================================

export interface PolicyConfig {
  quorum_required: number;
  total_approvers: number;
  timeout_hours: number;
  delegate_allowed: boolean;
  required_roles?: string[];
  auto_reject_after?: number;
  priority: Priority;
  metadata?: Record<string, unknown>;
}

export interface ApprovalDefinition {
  id: UUID;
  name: string;
  description: string;
  category: string;
  policy_config: PolicyConfig;
  version: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  created_by: UUID;
}

// ============================================
// Request and Action Types
// ============================================

export interface ApprovalRequest {
  id: UUID;
  definition_id: UUID;
  requester_id: UUID;
  resource_type: string;
  resource_id: string;
  context_data: Record<string, unknown>;
  status: ApprovalStatus;
  current_step: number;
  priority: Priority;
  deadline: string;
  workflow_id: string;
  metadata?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface ApproverAssignment {
  id: UUID;
  request_id: UUID;
  approver_id: UUID;
  role: string;
  step_order: number;
  has_acted: boolean;
  acted_at?: string;
  action_type?: ActionType;
  assigned_at: string;
}

export interface ApprovalAction {
  id: UUID;
  request_id: UUID;
  actor_id: UUID;
  actor_type: 'user' | 'system' | 'service';
  action_type: ActionType;
  on_behalf_of?: UUID;
  step_number: number;
  comments: string;
  signature_hash: string;
  previous_hash: string;
  ip_address?: string;
  user_agent?: string;
  timestamp: string;
}

export interface Delegation {
  id: UUID;
  delegator_id: UUID;
  delegatee_id: UUID;
  definition_id?: UUID;
  scope?: Record<string, unknown>;
  valid_from: string;
  valid_until: string;
  is_revoked: boolean;
  reason?: string;
  created_at: string;
  revoked_at?: string;
}

// ============================================
// Status and Summary Types
// ============================================

export interface ApprovalSummary {
  request_id: UUID;
  total_required: number;
  approvals_count: number;
  rejections_count: number;
  pending_count: number;
  quorum_met: boolean;
  status: ApprovalStatus;
  deadline: string;
  time_remaining: number; // in milliseconds
}

export interface AuditTrailEntry {
  action_id: UUID;
  actor_id: UUID;
  actor_name: string;
  action_type: ActionType;
  comments: string;
  timestamp: string;
  signature_valid: boolean;
  delegated_from?: UUID;
  ip_address?: string;
}

export interface AuditVerificationResult {
  request_id: UUID;
  actions_count: number;
  hash_chain_valid: boolean;
  broken_at_action?: UUID;
  verification_message: string;
}

export interface AuditReport {
  report_id: UUID;
  request_id: UUID;
  generated_at: string;
  generated_by: UUID;
  chain_status: ChainStatus;
  actions: AuditActionDetail[];
  compliance_notes: string[];
  summary: AuditSummary;
}

export interface ChainStatus {
  is_valid: boolean;
  total_entries: number;
  valid_entries: number;
  broken_at?: string[];
  first_entry_time: string;
  last_entry_time: string;
}

export interface AuditActionDetail {
  sequence_num: number;
  timestamp: string;
  actor_id: UUID;
  action_type: ActionType;
  comments: string;
  signature_valid: boolean;
  ip_address?: string;
  delegated_from?: UUID;
}

export interface AuditSummary {
  total_actions: number;
  approvals: number;
  rejections: number;
  abstentions: number;
  delegated_count: number;
  duration: string;
}

// ============================================
// API Request/Response Types
// ============================================

export interface CreateDefinitionRequest {
  name: string;
  description?: string;
  category: string;
  policy_config: PolicyConfig;
  created_by: string;
}

export interface CreateRequestRequest {
  definition_id: string;
  requester_id: string;
  resource_type: string;
  resource_id: string;
  context_data?: Record<string, unknown>;
  approver_ids: string[];
  priority?: Priority;
}

export interface DecisionRequest {
  approver_id: string;
  decision: ActionType;
  comments?: string;
  delegated_by?: string;
}

export interface CreateDelegationRequest {
  delegator_id: string;
  delegatee_id: string;
  definition_id?: string;
  scope?: Record<string, unknown>;
  valid_from: string;
  valid_until: string;
  reason?: string;
}

export interface PaginationParams {
  page?: number;
  limit?: number;
  sort_by?: string;
  sort_order?: 'asc' | 'desc';
}

export interface FilterParams extends PaginationParams {
  status?: ApprovalStatus[];
  category?: string;
  priority?: Priority[];
  from_date?: string;
  to_date?: string;
}

// ============================================
// Frontend Component Types
// ============================================

export interface ApprovalInboxItem {
  request: ApprovalRequest;
  summary: ApprovalSummary;
  is_overdue: boolean;
  assigned_to_me: boolean;
  can_act: boolean;
}

export interface ApprovalDetailState {
  request: ApprovalRequest;
  approvers: ApproverAssignment[];
  actions: ApprovalAction[];
  summary: ApprovalSummary;
  audit_trail: AuditTrailEntry[];
  selected_action?: ActionType;
  comments?: string;
  is_loading: boolean;
  is_submitting: boolean;
  error?: string;
}

export interface QuorumProgress {
  required: number;
  current: number;
  percentage: number;
  approvers: {
    id: UUID;
    name: string;
    has_acted: boolean;
    action?: ActionType;
    acted_at?: string;
  }[];
}

export interface ConfirmationDialogProps {
  is_open: boolean;
  title: string;
  message: string;
  action_type: ActionType;
  request_info: {
    id: UUID;
    resource_type: string;
    resource_id: string;
  };
  on_confirm: (comments: string) => void;
  on_cancel: () => void;
}

export interface NotificationPayload {
  request_id: UUID;
  recipient_id: UUID;
  notification_type: 'approval_required' | 'approved' | 'rejected' | 'expired';
  channel: NotificationChannel;
  data: Record<string, unknown>;
}

// ============================================
// Temporal Workflow Types
// ============================================

export interface WorkflowInput {
  request_id: string;
  definition_id: string;
  requester_id: string;
  resource_type: string;
  resource_id: string;
  context_data: Record<string, unknown>;
  approver_ids: string[];
  policy_config: PolicyConfig;
  priority: Priority;
  created_by: string;
}

export interface WorkflowOutput {
  request_id: string;
  final_status: ApprovalStatus;
  success: boolean;
  message: string;
  execution_time: number;
  approvals_received: number;
  rejections_received: number;
  actions_executed: string[];
  final_decision: string;
}

export interface ApprovalSignal {
  approver_id: string;
  action_type: ActionType;
  comments: string;
  delegated_by?: string;
  timestamp: string;
}

// ============================================
// Error Types
// ============================================

export interface ApiError {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export interface ValidationError {
  field: string;
  message: string;
  value?: unknown;
}

// ============================================
// Utility Types
// ============================================

export type AsyncState<T> = 
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: ApiError };

export type ActionHandler = (requestId: string, decision: ActionType, comments?: string) => Promise<void>;
