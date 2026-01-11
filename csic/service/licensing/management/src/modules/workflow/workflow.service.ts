import { Injectable } from '@nestjs/common';
import { ApplicationStatus } from '../applications/entities/application.entity';
import { LicenseStatus } from '../licenses/entities/license-status.enum';
import {
  APPLICATION_TRANSITIONS,
  LICENSE_TRANSITIONS,
  APPLICATION_STATE_METADATA,
  LICENSE_STATE_METADATA,
  StateMetadata,
} from './states';

@Injectable()
export class WorkflowService {
  /**
   * Check if a transition from one application status to another is allowed
   */
  canTransition(
    currentStatus: ApplicationStatus,
    targetStatus: ApplicationStatus,
  ): boolean {
    const allowedTransitions = APPLICATION_TRANSITIONS[currentStatus];
    if (!allowedTransitions) {
      return false;
    }
    return allowedTransitions.includes(targetStatus);
  }

  /**
   * Check if a transition from one license status to another is allowed
   */
  canTransitionLicense(
    currentStatus: LicenseStatus,
    targetStatus: LicenseStatus,
  ): boolean {
    const allowedTransitions = LICENSE_TRANSITIONS[currentStatus];
    if (!allowedTransitions) {
      return false;
    }
    return allowedTransitions.includes(targetStatus);
  }

  /**
   * Get all allowed transitions from a given application status
   */
  getAllowedTransitions(status: ApplicationStatus): ApplicationStatus[] {
    return APPLICATION_TRANSITIONS[status] || [];
  }

  /**
   * Get all allowed transitions from a given license status
   */
  getAllowedTransitionsLicense(status: LicenseStatus): LicenseStatus[] {
    return LICENSE_TRANSITIONS[status] || [];
  }

  /**
   * Get state metadata for an application status
   */
  getState(status: ApplicationStatus): StateMetadata {
    return APPLICATION_STATE_METADATA[status];
  }

  /**
   * Get state metadata for a license status
   */
  getLicenseState(status: LicenseStatus): StateMetadata {
    return LICENSE_STATE_METADATA[status];
  }

  /**
   * Get the next valid statuses in the workflow
   */
  getNextStatuses(status: ApplicationStatus): ApplicationStatus[] {
    return this.getAllowedTransitions(status);
  }

  /**
   * Check if a status is terminal (no further transitions possible)
   */
  isTerminalStatus(status: ApplicationStatus): boolean {
    const transitions = APPLICATION_TRANSITIONS[status];
    return !transitions || transitions.length === 0;
  }

  /**
   * Check if a license status is terminal
   */
  isTerminalLicenseStatus(status: LicenseStatus): boolean {
    const transitions = LICENSE_TRANSITIONS[status];
    return !transitions || transitions.length === 0;
  }

  /**
   * Get workflow progress percentage
   */
  getProgressPercentage(status: ApplicationStatus): number {
    const workflowOrder = [
      ApplicationStatus.DRAFT,
      ApplicationStatus.SUBMITTED,
      ApplicationStatus.UNDER_REVIEW,
      ApplicationStatus.APPROVED,
      ApplicationStatus.RENEWAL_WINDOW,
      ApplicationStatus.RENEWAL_IN_PROGRESS,
    ];

    const terminalStatuses = [
      ApplicationStatus.REJECTED,
      ApplicationStatus.EXPIRED,
      ApplicationStatus.REVOKED,
    ];

    // Terminal states have 100% or 0% depending on outcome
    if (terminalStatuses.includes(status)) {
      return status === ApplicationStatus.REJECTED ? 0 : 100;
    }

    const index = workflowOrder.indexOf(status);
    if (index === -1) return 0;

    return ((index + 1) / workflowOrder.length) * 100;
  }

  /**
   * Validate a complete workflow transition with full context
   */
  validateTransition(
    currentStatus: ApplicationStatus,
    targetStatus: ApplicationStatus,
    context?: {
      hasRequiredDocuments?: boolean;
      hasReviewNotes?: boolean;
      isOverdue?: boolean;
    },
  ): { valid: boolean; reason?: string } {
    // Basic transition check
    if (!this.canTransition(currentStatus, targetStatus)) {
      return {
        valid: false,
        reason: `Cannot transition from ${currentStatus} to ${targetStatus}`,
      };
    }

    // Context-specific validations
    if (targetStatus === ApplicationStatus.SUBMITTED) {
      if (!context?.hasRequiredDocuments) {
        return {
          valid: false,
          reason: 'Cannot submit: missing required documents',
        };
      }
    }

    if (targetStatus === ApplicationStatus.APPROVED || targetStatus === ApplicationStatus.REJECTED) {
      if (!context?.hasReviewNotes) {
        return {
          valid: false,
          reason: 'Cannot make decision: no review notes recorded',
        };
      }
    }

    return { valid: true };
  }

  /**
   * Get a summary of the workflow for debugging/monitoring
   */
  getWorkflowSummary(): any {
    return Object.entries(APPLICATION_TRANSITIONS).map(([status, transitions]) => ({
      status,
      isTerminal: transitions.length === 0,
      nextStatuses: transitions,
      metadata: APPLICATION_STATE_METADATA[status as ApplicationStatus],
    }));
  }
}
