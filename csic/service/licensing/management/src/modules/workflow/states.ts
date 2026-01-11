import { ApplicationStatus } from '../applications/entities/application.entity';
import { LicenseStatus } from '../licenses/entities/license-status.enum';

// Application State Machine Transitions
export const APPLICATION_TRANSITIONS: Record<ApplicationStatus, ApplicationStatus[]> = {
  [ApplicationStatus.DRAFT]: [ApplicationStatus.SUBMITTED],
  [ApplicationStatus.SUBMITTED]: [ApplicationStatus.UNDER_REVIEW, ApplicationStatus.PENDING_INFO],
  [ApplicationStatus.UNDER_REVIEW]: [
    ApplicationStatus.APPROVED,
    ApplicationStatus.REJECTED,
    ApplicationStatus.PENDING_INFO,
  ],
  [ApplicationStatus.PENDING_INFO]: [ApplicationStatus.SUBMITTED, ApplicationStatus.UNDER_REVIEW],
  [ApplicationStatus.APPROVED]: [ApplicationStatus.RENEWAL_WINDOW],
  [ApplicationStatus.REJECTED]: [],
  [ApplicationStatus.RENEWAL_WINDOW]: [ApplicationStatus.RENEWAL_IN_PROGRESS],
  [ApplicationStatus.RENEWAL_IN_PROGRESS]: [ApplicationStatus.APPROVED],
  [ApplicationStatus.SUSPENDED]: [ApplicationStatus.EXPIRED],
  [ApplicationStatus.EXPIRED]: [],
  [ApplicationStatus.REVOKED]: [],
};

// License State Machine Transitions
export const LICENSE_TRANSITIONS: Record<LicenseStatus, LicenseStatus[]> = {
  [LicenseStatus.ACTIVE]: [LicenseStatus.SUSPENDED, LicenseStatus.PENDING_RENEWAL],
  [LicenseStatus.SUSPENDED]: [LicenseStatus.ACTIVE, LicenseStatus.REVOKED, LicenseStatus.EXPIRED],
  [LicenseStatus.EXPIRED]: [LicenseStatus.REVOKED],
  [LicenseStatus.REVOKED]: [],
  [LicenseStatus.PENDING_RENEWAL]: [LicenseStatus.ACTIVE, LicenseStatus.EXPIRED],
};

// State metadata for UI and documentation
export const APPLICATION_STATE_METADATA: Record<ApplicationStatus, StateMetadata> = {
  [ApplicationStatus.DRAFT]: {
    name: 'Draft',
    description: 'Application is being prepared by the entity',
    color: '#6B7280',
    icon: 'file-edit',
  },
  [ApplicationStatus.SUBMITTED]: {
    name: 'Submitted',
    description: 'Application submitted and awaiting review',
    color: '#3B82F6',
    icon: 'send',
  },
  [ApplicationStatus.UNDER_REVIEW]: {
    name: 'Under Review',
    description: 'Regulator is actively reviewing the application',
    color: '#F59E0B',
    icon: 'search',
  },
  [ApplicationStatus.PENDING_INFO]: {
    name: 'Pending Information',
    description: 'Additional information requested from applicant',
    color: '#EF4444',
    icon: 'alert-circle',
  },
  [ApplicationStatus.APPROVED]: {
    name: 'Approved',
    description: 'Application approved, license to be issued',
    color: '#10B981',
    icon: 'check-circle',
  },
  [ApplicationStatus.REJECTED]: {
    name: 'Rejected',
    description: 'Application rejected',
    color: '#DC2626',
    icon: 'x-circle',
  },
  [ApplicationStatus.RENEWAL_WINDOW]: {
    name: 'Renewal Window',
    description: 'License renewal period open',
    color: '#8B5CF6',
    icon: 'clock',
  },
  [ApplicationStatus.RENEWAL_IN_PROGRESS]: {
    name: 'Renewal In Progress',
    description: 'Renewal application under review',
    color: '#6366F1',
    icon: 'refresh-cw',
  },
  [ApplicationStatus.SUSPENDED]: {
    name: 'Suspended',
    description: 'License temporarily suspended',
    color: '#F97316',
    icon: 'pause-circle',
  },
  [ApplicationStatus.EXPIRED]: {
    name: 'Expired',
    description: 'License has expired',
    color: '#991B1B',
    icon: 'calendar-x',
  },
  [ApplicationStatus.REVOKED]: {
    name: 'Revoked',
    description: 'License permanently revoked',
    color: '#7F1D1D',
    icon: 'ban',
  },
};

export const LICENSE_STATE_METADATA: Record<LicenseStatus, StateMetadata> = {
  [LicenseStatus.ACTIVE]: {
    name: 'Active',
    description: 'License is valid and in effect',
    color: '#10B981',
    icon: 'check-circle',
  },
  [LicenseStatus.SUSPENDED]: {
    name: 'Suspended',
    description: 'License temporarily suspended',
    color: '#F97316',
    icon: 'pause-circle',
  },
  [LicenseStatus.EXPIRED]: {
    name: 'Expired',
    description: 'License has expired',
    color: '#991B1B',
    icon: 'calendar-x',
  },
  [LicenseStatus.REVOKED]: {
    name: 'Revoked',
    description: 'License permanently revoked',
    color: '#7F1D1D',
    icon: 'ban',
  },
  [LicenseStatus.PENDING_RENEWAL]: {
    name: 'Pending Renewal',
    description: 'Renewal pending approval',
    color: '#6366F1',
    icon: 'clock',
  },
};

export interface StateMetadata {
  name: string;
  description: string;
  color: string;
  icon: string;
}
