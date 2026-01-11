// Frontend Component Exports
// Export all approval system components for easy importing

export { ApprovalInbox } from './ApprovalInbox';
export { ApprovalDetail } from './ApprovalDetail';
export { ConfirmationDialog } from './ConfirmationDialog';

// Re-export types
export * from '../types/approval';

// Re-export hooks
export { useApproval, useApprovalRequest, useAudit } from '../hooks/useApproval';

// Re-export API service
export { approvalApi } from '../services/approvalApi';
