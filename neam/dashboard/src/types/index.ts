/**
 * NEAM Government Decision Dashboards - TypeScript Type Definitions
 * Comprehensive type definitions for all dashboard features
 */

// User and Authentication Types
export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  region?: string;
  department?: string;
  permissions: Permission[];
  avatar?: string;
  createdAt: Date;
  lastLogin?: Date;
}

export type UserRole = 'PMO' | 'FINANCE' | 'STATE_ADMIN' | 'DISTRICT_COLLECTOR' | 'ANALYST';

export type Permission =
  | 'VIEW_DASHBOARD'
  | 'VIEW_NATIONAL'
  | 'VIEW_REGIONAL'
  | 'VIEW_DISTRICT'
  | 'CREATE_POLICY'
  | 'EDIT_POLICY'
  | 'DELETE_POLICY'
  | 'APPROVE_INTERVENTION'
  | 'TRIGGER_INTERVENTION'
  | 'CANCEL_INTERVENTION'
  | 'CREATE_SIMULATION'
  | 'VIEW_ANALYTICS'
  | 'EXPORT_DATA'
  | 'MANAGE_USERS'
  | 'VIEW_REPORTS'
  | 'GENERATE_REPORTS';

export interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
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
  expiresIn: number;
}

// Dashboard Data Types
export interface DashboardSummary {
  totalMetrics: number;
  activeAlerts: number;
  activeInterventions: number;
  regionsMonitored: number;
  lastUpdated: Date;
  systemHealth: 'healthy' | 'warning' | 'critical';
  overview: {
    totalGDP: number;
    gdpGrowthRate: number;
    inflationRate: number;
    unemploymentRate: number;
    fiscalDeficit: number;
    tradeBalance: number;
  };
  keyMetrics: MetricCard[];
  regionalPerformance: RegionPerformance[];
  sectorPerformance: SectorPerformance[];
}

export interface MetricCard {
  id: string;
  name: string;
  value: number;
  unit: string;
  change: number;
  changeType: 'increase' | 'decrease' | 'neutral';
  trend: TrendData[];
  threshold?: Threshold;
  status: 'normal' | 'warning' | 'critical';
  category: MetricCategory;
  description?: string;
}

export type MetricCategory =
  | 'GDP'
  | 'INFLATION'
  | 'EMPLOYMENT'
  | 'AGRICULTURE'
  | 'MANUFACTURING'
  | 'SERVICES'
  | 'INFRASTRUCTURE'
  | 'TRADE'
  | 'REVENUE'
  | 'EXPENDITURE';

export interface TrendData {
  timestamp: Date;
  value: number;
}

export interface Threshold {
  warning: number;
  critical: number;
  label: string;
}

// Regional Data Types
export interface RegionData {
  id: string;
  code: string;
  name: string;
  state: string;
  stressIndex: number;
  status: RegionStatus;
  metrics: RegionalMetrics;
  activeAlerts: number;
  activeInterventions: number;
  population?: number;
  gdp?: number;
  type: 'NATIONAL' | 'STATE' | 'DISTRICT';
  parentId?: string;
  coordinates?: GeoCoordinates;
}

export interface GeoCoordinates {
  latitude: number;
  longitude: number;
}

export type RegionStatus = 'healthy' | 'caution' | 'warning' | 'critical';

export interface RegionalMetrics {
  inflationRate: number;
  unemploymentRate: number;
  gdpGrowth: number;
  agriculturalOutput: number;
  industrialOutput: number;
  consumerSpending: number;
  revenueCollection: number;
  expenditure: number;
}

export interface RegionPerformance {
  regionId: string;
  regionName: string;
  performanceIndex: number;
  rank: number;
  changeFromLastPeriod: number;
  indicators: MetricCard[];
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
}

export interface SectorPerformance {
  sectorId: string;
  sectorName: string;
  contributionToGDP: number;
  growthRate: number;
  employmentShare: number;
  investmentRate: number;
  trend: 'rising' | 'falling' | 'stable';
  subSectors: SubSector[];
}

export interface SubSector {
  id: string;
  name: string;
  value: number;
  growth: number;
  share: number;
}

// Heatmap Types
export interface HeatmapCell {
  regionId: string;
  regionName: string;
  x: number;
  y: number;
  value: number;
  status: RegionStatus;
  alertCount: number;
  interventionCount: number;
  code?: string;
}

export interface HeatmapConfig {
  metric: string;
  colorScale: ColorScale;
  showLegend: boolean;
  showTooltips: boolean;
  interactive: boolean;
  projection?: string;
}

export interface ColorScale {
  low: string;
  medium: string;
  high: string;
  critical: string;
  steps?: number;
}

// Alert Types
export interface Alert {
  id: string;
  type: AlertType;
  priority: AlertPriority;
  title: string;
  message: string;
  regionId?: string;
  regionName?: string;
  sectorId?: string;
  policyId?: string;
  interventionId?: string;
  createdAt: Date;
  acknowledgedAt?: Date;
  acknowledgedBy?: string;
  resolvedAt?: Date;
  metadata?: Record<string, unknown>;
}

export type AlertType =
  | 'INFLATION'
  | 'EMPLOYMENT'
  | 'AGRICULTURE'
  | 'TRADE'
  | 'ENERGY'
  | 'FINANCIAL'
  | 'EMERGENCY'
  | 'THRESHOLD_BREACH'
  | 'TREND_REVERSAL'
  | 'DATA_ANOMALY'
  | 'REGIONAL_DISPARITY';

export type AlertPriority = 'low' | 'medium' | 'high' | 'critical';

// Intervention Types
export interface Intervention {
  id: string;
  policyId: string;
  policyName: string;
  workflowId: string;
  status: InterventionStatus;
  regionId: string;
  regionName?: string;
  sectorId?: string;
  triggeredAt: Date;
  completedAt?: Date;
  progress: number;
  currentStep: string;
  outcome?: InterventionOutcome;
  type: InterventionType;
  name: string;
  description: string;
  budgetAllocated: number;
  budgetUtilized: number;
  targetRegions: string[];
  targetSectors: string[];
  expectedOutcome: string;
  actualOutcome?: number;
  kpis: InterventionKPI[];
}

export type InterventionType =
  | 'FISCAL_STIMULUS'
  | 'TAX_RELIEF'
  | 'SUBSIDY_PROGRAM'
  | 'INFRASTRUCTURE_PROJECT'
  | 'EMPLOYMENT_SCHEME'
  | 'TRAINING_PROGRAM'
  | 'REGULATORY_REFORM';

export type InterventionStatus =
  | 'PENDING'
  | 'APPROVED'
  | 'EXECUTING'
  | 'COMPLETED'
  | 'CANCELLED'
  | 'FAILED'
  | 'ON_HOLD';

export interface InterventionOutcome {
  targetAchieved: boolean;
  impactScore: number;
  preMetrics: Record<string, number>;
  postMetrics: Record<string, number>;
  beneficiaries: number;
  costIncurred: number;
}

export interface InterventionKPI {
  id: string;
  name: string;
  target: number;
  current: number;
  unit: string;
  progress: number;
}

// Policy Types
export interface Policy {
  id: string;
  name: string;
  description: string;
  category: PolicyCategory;
  priority: number;
  status: PolicyStatus;
  rules: PolicyRules;
  targetScope: TargetScope;
  actions: PolicyAction[];
  createdAt: Date;
  updatedAt: Date;
  version: number;
  createdBy: string;
}

export type PolicyCategory =
  | 'INFLATION'
  | 'EMPLOYMENT'
  | 'AGRICULTURE'
  | 'MANUFACTURING'
  | 'TRADE'
  | 'ENERGY'
  | 'FINANCIAL'
  | 'EMERGENCY';

export type PolicyStatus = 'DRAFT' | 'ACTIVE' | 'INACTIVE' | 'ARCHIVED';

export interface PolicyRules {
  conditions: RuleCondition[];
  logicExpression: string;
  triggerType: 'IMMEDIATE' | 'SUSTAINED' | 'SCHEDULED';
  thresholdValues: Record<string, number>;
  timeWindow?: number;
}

export interface RuleCondition {
  metric: string;
  operator: 'GT' | 'LT' | 'EQ' | 'GTE' | 'LTE';
  value: number;
}

export interface TargetScope {
  regions: string[];
  sectors: string[];
}

export interface PolicyAction {
  type: 'NOTIFY' | 'ALERT' | 'WORKFLOW' | 'STATE_CHANGE' | 'RESTRICTION';
  parameters: Record<string, unknown>;
}

// Simulation Types
export interface Simulation {
  id: string;
  name: string;
  description: string;
  status: SimulationStatus;
  createdBy: string;
  createdAt: Date;
  completedAt?: Date;
  baseScenario: BaseScenario;
  interventions: SimulationIntervention[];
  results?: SimulationResult;
}

export type SimulationStatus = 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED' | 'CANCELLED';

export interface BaseScenario {
  regionId: string;
  sectorId?: string;
  startDate: Date;
  endDate: Date;
  initialState: Record<string, number>;
}

export interface SimulationIntervention {
  type: string;
  parameters: Record<string, unknown>;
  startDate: Date;
  duration: number;
}

export interface SimulationResult {
  projectedMetrics: Record<string, TimeSeriesPoint[]>;
  baselineMetrics: Record<string, TimeSeriesPoint[]>;
  comparison: Record<string, ComparisonMetric>;
  summary: SimulationSummary;
  confidence: number;
}

export interface TimeSeriesPoint {
  timestamp: string;
  value: number;
}

export interface ComparisonMetric {
  baselineValue: number;
  simulatedValue: number;
  delta: number;
  deltaPercent: number;
  impact: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
}

export interface SimulationSummary {
  overallImpact: 'POSITIVE' | 'NEGATIVE' | 'MIXED';
  primaryBenefit: string;
  primaryCost: string;
  netPresentValue: number;
  roi: number;
}

// Analytics Types
export interface AnalyticsData {
  period: 'day' | 'week' | 'month' | 'quarter' | 'year';
  startDate: Date;
  endDate: Date;
  metrics: AnalyticsMetric[];
  comparisons: AnalyticsComparison[];
}

export interface AnalyticsMetric {
  id: string;
  name: string;
  currentValue: number;
  previousValue: number;
  change: number;
  changePercent: number;
  trend: TrendDirection;
  timeSeries: TimeSeriesPoint[];
}

export type TrendDirection = 'UP' | 'DOWN' | 'STABLE';

export interface AnalyticsComparison {
  metricId: string;
  regions: RegionComparison[];
}

export interface RegionComparison {
  regionId: string;
  regionName: string;
  value: number;
  rank: number;
}

export interface DrillDownPath {
  path: RegionData[];
  currentLevel: number;
  maxLevel: number;
  canDrillUp: boolean;
  canDrillDown: boolean;
}

export interface ComparativeAnalysis {
  baseRegion: RegionData;
  comparisonRegions: RegionData[];
  metrics: ComparativeMetric[];
  insights: AnalysisInsight[];
}

export interface ComparativeMetric {
  indicator: string;
  baseValue: number;
  comparisonValues: Record<string, number>;
  variance: Record<string, number>;
  ranking: RegionRanking[];
}

export interface RegionRanking {
  regionId: string;
  regionName: string;
  value: number;
  rank: number;
}

export interface AnalysisInsight {
  type: 'outlier' | 'trend' | 'correlation' | 'forecast' | 'recommendation';
  title: string;
  description: string;
  confidence: number;
  impactedRegions?: string[];
  relatedIndicators?: string[];
}

// Export Types
export interface ExportConfig {
  format: 'pdf' | 'csv' | 'xlsx';
  dateRange: { start: Date; end: Date };
  sections: ExportSection[];
  includeCharts: boolean;
  includeTables: boolean;
  title?: string;
  watermark?: string;
  confidentialityLevel?: string;
}

export type ExportSection =
  | 'SUMMARY'
  | 'METRICS'
  | 'ALERTS'
  | 'INTERVENTIONS'
  | 'REGIONAL_DATA'
  | 'POLICIES'
  | 'ANALYTICS'
  | 'COMPARISONS';

export interface ExportJob {
  id: string;
  config: ExportConfig;
  status: ExportStatus;
  progress: number;
  createdAt: Date;
  startedAt?: Date;
  completedAt?: Date;
  error?: string;
  downloadUrl?: string;
}

export type ExportStatus = 'queued' | 'processing' | 'completed' | 'failed' | 'cancelled';

export interface ReportConfig {
  reportId?: string;
  title: string;
  type: ReportType;
  format: 'pdf' | 'csv' | 'xlsx';
  sections: ReportSection[];
  filters: DashboardFilters;
  includeCharts: boolean;
  includeTables: boolean;
  watermark?: string;
  confidentialityLevel: 'public' | 'restricted' | 'confidential' | 'secret';
  generatedBy: string;
  generatedAt: Date;
}

export type ReportType =
  | 'PERFORMANCE_SUMMARY'
  | 'REGIONAL_ANALYSIS'
  | 'SECTOR_BREAKDOWN'
  | 'ALERT_LOG'
  | 'COMPARATIVE_STUDY'
  | 'TREND_REPORT'
  | 'AUDIT_TRAIL'
  | 'CUSTOM';

export interface ReportSection {
  id: string;
  title: string;
  type: 'text' | 'chart' | 'table' | 'metric' | 'map';
  config: Record<string, unknown>;
  order: number;
}

// Activity Types
export interface Activity {
  id: string;
  type: ActivityType;
  title: string;
  description: string;
  userId: string;
  userName: string;
  regionId?: string;
  timestamp: Date;
  metadata?: Record<string, unknown>;
}

export type ActivityType =
  | 'DATA_UPDATE'
  | 'REPORT_GENERATED'
  | 'ALERT_TRIGGERED'
  | 'POLICY_APPROVED'
  | 'INTERVENTION_STARTED'
  | 'CONFIGURATION_CHANGED';

// API Response Types
export interface ApiResponse<T> {
  data: T;
  success: boolean;
  message?: string;
  errors?: ApiError[];
  pagination?: Pagination;
}

export interface ApiError {
  code: string;
  message: string;
  field?: string;
}

export interface Pagination {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

// Filter Types
export interface DashboardFilters {
  dateRange?: DateRange;
  regions?: string[];
  sectors?: string[];
  alertTypes?: AlertType[];
  statuses?: string[];
  priorities?: AlertPriority[];
  indicators?: string[];
  comparisonMode?: 'period' | 'region' | 'sector';
}

export interface DateRange {
  start: Date;
  end: Date;
  preset?: DatePreset;
}

export type DatePreset =
  | 'TODAY'
  | 'WEEK'
  | 'MONTH'
  | 'QUARTER'
  | 'YEAR'
  | 'FY'
  | 'CUSTOM';

// Notification Types
export interface Notification {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message: string;
  read: boolean;
  createdAt: Date;
  actionUrl?: string;
  duration?: number;
}

// UI State Types
export interface UIState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  loading: Record<string, boolean>;
  modals: Record<string, boolean>;
  toasts: Notification[];
  selectedRegion?: string;
  selectedSector?: string;
  activeView: 'overview' | 'analytics' | 'reports' | 'interventions' | 'settings';
}

// Map Visualization Types
export interface MapConfig {
  projection: 'mercator' | 'orthographic' | 'conicConformal';
  center: [number, number];
  scale: number;
  colors: ColorScale;
  legend: MapLegend;
}

export interface MapLegend {
  title: string;
  orientation: 'horizontal' | 'vertical';
  position: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right';
}

export interface MapDataPoint {
  regionId: string;
  value: number;
  label: string;
  tooltip?: string;
}

// Chart Configuration Types
export interface ChartConfig {
  type: 'line' | 'bar' | 'area' | 'pie' | 'scatter' | 'composed' | 'radar';
  title?: string;
  xAxis?: ChartAxis;
  yAxis?: ChartAxis;
  legend?: ChartLegend;
  tooltip?: ChartTooltip;
  colors?: string[];
  animate?: boolean;
}

export interface ChartAxis {
  label?: string;
  dataKey: string;
  type?: 'category' | 'number' | 'time';
  tickFormatter?: (value: unknown) => string;
}

export interface ChartLegend {
  position: 'top' | 'bottom' | 'left' | 'right';
  align?: 'left' | 'center' | 'right';
}

export interface ChartTooltip {
  enabled: boolean;
  formatter?: (value: unknown, name: string) => string;
}

// Trend Comparison Types
export interface TrendComparison {
  currentPeriod: string;
  previousPeriod: string;
  metrics: ComparisonMetric[];
}

export interface ComparisonMetric {
  indicatorCode: string;
  indicatorName: string;
  currentValue: number;
  previousValue: number;
  change: number;
  changePercent: number;
  trend: TrendDirection;
}
