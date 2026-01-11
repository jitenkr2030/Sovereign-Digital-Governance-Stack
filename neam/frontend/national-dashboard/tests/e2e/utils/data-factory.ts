import { Faker } from '@faker-js/faker';
import { v4 as uuidv4 } from 'uuid';

/**
 * Test Data Factory
 * 
 * Provides utilities for generating test data across all E2E workflows.
 * Uses Faker.js for realistic data generation.
 */

// Initialize Faker with a fixed seed for reproducibility
export const faker = new Faker({
  locale: ['en', 'en_GB'],
});

// Test run identifier for data isolation
export const getTestRunId = (): string => process.env.TEST_RUN_ID || `test-${Date.now()}`;

// Tenant ID for multi-tenant isolation
export const getTenantId = (): string => {
  return process.env.TEST_TENANT_ID || `test-tenant-${getTestRunId().slice(0, 8)}`;
};

/**
 * Generate a unique identifier with test run prefix
 */
export const uniqueId = (prefix: string = 'entity'): string => {
  return `${prefix}-${getTestRunId().slice(0, 8)}-${uuidv4().slice(0, 8)}`;
};

/**
 * Data Generation Factory
 */
export const DataFactory = {
  /**
   * Generate raw economic data for ingestion
   */
  createRawEconomicData: (overrides?: Partial<EconomicData>): EconomicData => {
    const regions = ['Northeast', 'Southeast', 'Midwest', 'Southwest', 'West'];
    const dataTypes = ['GDP', 'Employment', 'Inflation', 'Trade', 'Investment'];
    
    return {
      id: uniqueId('raw-data'),
      tenantId: getTenantId(),
      timestamp: faker.date.recent({ days: 7 }).toISOString(),
      region: overrides?.region || faker.helpers.arrayElement(regions),
      dataType: overrides?.dataType || faker.helpers.arrayElement(dataTypes),
      value: overrides?.value ?? faker.number.float({ min: -10, max: 15, fractionDigits: 2 }),
      unit: overrides?.unit || '%',
      source: overrides?.source || faker.company.name(),
      metadata: {
        collectionMethod: faker.helpers.arrayElement(['survey', 'automated', 'manual']),
        confidenceLevel: faker.number.float({ min: 0.8, max: 0.99, fractionDigits: 2 }),
        qualityScore: faker.number.int({ min: 70, max: 100 }),
      },
      ...overrides,
    };
  },

  /**
   * Generate intelligence item from analyzed data
   */
  createIntelligenceItem: (overrides?: Partial<IntelligenceItem>): IntelligenceItem => {
    const severities = ['info', 'low', 'medium', 'high', 'critical'];
    const categories = ['Economic Trend', 'Market Anomaly', 'Policy Impact', 'Risk Indicator'];
    const statuses = ['new', 'analyzing', 'identified', 'resolved'];
    
    return {
      id: uniqueId('intel'),
      tenantId: getTenantId(),
      title: overrides?.title || `Intelligence Alert: ${faker.commerce.productName()}`,
      description: overrides?.description || faker.lorem.paragraph(2),
      severity: overrides?.severity || faker.helpers.arrayElement(severities),
      category: overrides?.category || faker.helpers.arrayElement(categories),
      status: overrides?.status || 'new',
      sourceDataId: uniqueId('source'),
      analysis: {
        summary: faker.lorem.sentence(),
        confidence: faker.number.float({ min: 0.7, max: 0.99, fractionDigits: 2 }),
        recommendations: [
          faker.lorem.sentence(),
          faker.lorem.sentence(),
        ],
      },
      affectedRegions: overrides?.affectedRegions || [faker.helpers.arrayElement(['Northeast', 'Southeast', 'Midwest'])],
      createdAt: faker.date.recent({ days: 3 }).toISOString(),
      updatedAt: new Date().toISOString(),
      createdBy: 'system',
      ...overrides,
    };
  },

  /**
   * Generate intervention action
   */
  createIntervention: (overrides?: Partial<Intervention>): Intervention => {
    const types = ['monetary', 'fiscal', 'regulatory', 'emergency'];
    const statuses = ['draft', 'pending_approval', 'approved', 'active', 'completed', 'cancelled'];
    const priorities = ['low', 'medium', 'high', 'critical'];
    
    return {
      id: uniqueId('intervention'),
      tenantId: getTenantId(),
      name: overrides?.name || `Intervention: ${faker.commerce.productName()}`,
      description: faker.lorem.paragraph(2),
      type: overrides?.type || faker.helpers.arrayElement(types),
      status: overrides?.status || 'draft',
      priority: faker.helpers.arrayElement(priorities),
      targetIntelligenceId: uniqueId('intel'),
      parameters: {
        budget: faker.number.int({ min: 100000, max: 10000000 }),
        duration: faker.number.int({ min: 7, max: 365 }),
        resources: Array.from({ length: 3 }, () => faker.commerce.department()),
      },
      expectedOutcome: faker.lorem.sentence(),
      riskAssessment: {
        level: faker.helpers.arrayElement(['low', 'medium', 'high']),
        mitigation: faker.lorem.sentence(),
      },
      approvals: [],
      timeline: [
        {
          status: 'created',
          timestamp: new Date().toISOString(),
          actor: 'system',
          notes: 'Intervention created',
        },
      ],
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      createdBy: overrides?.createdBy || 'analyst@neam.gov',
      ...overrides,
    };
  },

  /**
   * Generate approval record
   */
  createApproval: (overrides?: Partial<Approval>): Approval => {
    const statuses = ['pending', 'approved', 'rejected', 'escalated'];
    const roles = ['supervisor', 'director', 'executive'];
    
    return {
      id: uniqueId('approval'),
      tenantId: getTenantId(),
      interventionId: uniqueId('intervention'),
      status: overrides?.status || 'pending',
      approverRole: overrides?.approverRole || faker.helpers.arrayElement(roles),
      approverEmail: faker.internet.email(),
      comments: faker.lorem.sentence(),
      decisionAt: null,
      expiresAt: faker.date.future({ years: 1 }).toISOString(),
      ...overrides,
    };
  },

  /**
   * Generate report configuration
   */
  createReportConfig: (overrides?: Partial<ReportConfig>): ReportConfig => {
    const types = ['executive_summary', 'detailed_analysis', 'compliance', 'trend_analysis'];
    const formats = ['pdf', 'excel', 'csv'];
    
    return {
      id: uniqueId('report'),
      tenantId: getTenantId(),
      name: overrides?.name || `Report: ${faker.commerce.productName()}`,
      type: overrides?.type || faker.helpers.arrayElement(types),
      format: faker.helpers.arrayElement(formats),
      dateRange: {
        start: faker.date.past({ years: 1 }).toISOString().split('T')[0],
        end: new Date().toISOString().split('T')[0],
      },
      filters: {
        regions: ['Northeast', 'Southeast'],
        dataTypes: ['GDP', 'Employment'],
        severity: ['high', 'critical'],
      },
      schedule: overrides?.schedule || null,
      createdAt: new Date().toISOString(),
      createdBy: 'analyst@neam.gov',
      ...overrides,
    };
  },

  /**
   * Generate user for authentication
   */
  createUser: (overrides?: Partial<User>): User => {
    const roles = ['super_admin', 'admin', 'analyst', 'viewer', 'supervisor'];
    
    return {
      id: uniqueId('user'),
      email: faker.internet.email(),
      firstName: faker.person.firstName(),
      lastName: faker.person.lastName(),
      role: overrides?.role || faker.helpers.arrayElement(roles),
      department: faker.commerce.department(),
      isActive: true,
      permissions: [],
      ...overrides,
    };
  },
};

// Type definitions for generated data
export interface EconomicData {
  id: string;
  tenantId: string;
  timestamp: string;
  region: string;
  dataType: string;
  value: number;
  unit: string;
  source: string;
  metadata: {
    collectionMethod: string;
    confidenceLevel: number;
    qualityScore: number;
  };
}

export interface IntelligenceItem {
  id: string;
  tenantId: string;
  title: string;
  description: string;
  severity: string;
  category: string;
  status: string;
  sourceDataId: string;
  analysis: {
    summary: string;
    confidence: number;
    recommendations: string[];
  };
  affectedRegions: string[];
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

export interface Intervention {
  id: string;
  tenantId: string;
  name: string;
  description: string;
  type: string;
  status: string;
  priority: string;
  targetIntelligenceId: string;
  parameters: {
    budget: number;
    duration: number;
    resources: string[];
  };
  expectedOutcome: string;
  riskAssessment: {
    level: string;
    mitigation: string;
  };
  approvals: Approval[];
  timeline: TimelineEvent[];
  createdAt: string;
  updatedAt: string;
  createdBy: string;
}

export interface Approval {
  id: string;
  tenantId: string;
  interventionId: string;
  status: string;
  approverRole: string;
  approverEmail: string;
  comments: string;
  decisionAt: string | null;
  expiresAt: string;
}

export interface TimelineEvent {
  status: string;
  timestamp: string;
  actor: string;
  notes: string;
}

export interface ReportConfig {
  id: string;
  tenantId: string;
  name: string;
  type: string;
  format: string;
  dateRange: {
    start: string;
    end: string;
  };
  filters: {
    regions: string[];
    dataTypes: string[];
    severity: string[];
  };
  schedule: string | null;
  createdAt: string;
  createdBy: string;
}

export interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  role: string;
  department: string;
  isActive: boolean;
  permissions: string[];
}

export default DataFactory;
