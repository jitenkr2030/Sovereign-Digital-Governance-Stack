/**
 * Authentication Constants
 * 
 * Defines roles, permissions, and configuration constants
 * for the role-based access control (RBAC) system.
 */

/**
 * User roles in the system
 */
export enum Role {
  SUPER_ADMIN = 'SuperAdmin',
  ADMIN = 'Admin',
  MANAGER = 'Manager',
  ANALYST = 'Analyst',
  OPERATOR = 'Operator',
  VIEWER = 'Viewer',
  GUEST = 'Guest',
}

/**
 * Role hierarchy (higher index = higher privilege)
 */
export const ROLE_HIERARCHY: Record<Role, number> = {
  [Role.SUPER_ADMIN]: 100,
  [Role.ADMIN]: 80,
  [Role.MANAGER]: 60,
  [Role.ANALYST]: 40,
  [Role.OPERATOR]: 30,
  [Role.VIEWER]: 20,
  [Role.GUEST]: 10,
};

/**
 * Default roles assigned to new users
 */
export const DEFAULT_USER_ROLES: Role[] = [Role.VIEWER];

/**
 * Admin roles that bypass permission checks
 */
export const ADMIN_ROLES: Role[] = [
  Role.SUPER_ADMIN,
  Role.ADMIN,
];

/**
 * Resource names for permission checks
 */
export enum Resource {
  DASHBOARD = 'dashboard',
  USERS = 'users',
  SETTINGS = 'settings',
  REPORTS = 'reports',
  ANALYTICS = 'analytics',
  AUDIT_LOGS = 'audit_logs',
  NETWORK_ASSETS = 'network_assets',
  INTERVENTIONS = 'interventions',
  POLICIES = 'policies',
  ALERTS = 'alerts,
  DATA_SOURCES = 'data_sources',
  EXPORTS = 'exports',
}

/**
 * Action types for permissions
 */
export enum Action {
  CREATE = 'create',
  READ = 'read',
  UPDATE = 'update',
  DELETE = 'delete',
  EXPORT = 'export',
  APPROVE = 'approve',
  EXECUTE = 'execute',
  ADMIN = 'admin',
}

/**
 * Role-based permission matrix
 * Defines which roles have access to which resources and actions
 */
export const ROLE_PERMISSIONS: Record<Role, Record<Resource, Action[]>> = {
  [Role.SUPER_ADMIN]: {
    [Resource.DASHBOARD]: [Action.READ, Action.EXECUTE],
    [Resource.USERS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.ADMIN],
    [Resource.SETTINGS]: [Action.READ, Action.UPDATE, Action.ADMIN],
    [Resource.REPORTS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.EXPORT],
    [Resource.ANALYTICS]: [Action.READ, Action.EXPORT],
    [Resource.AUDIT_LOGS]: [Action.READ],
    [Resource.NETWORK_ASSETS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.EXECUTE],
    [Resource.INTERVENTIONS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.APPROVE, Action.EXECUTE],
    [Resource.POLICIES]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.APPROVE],
    [Resource.ALERTS]: [Action.READ, Action.UPDATE],
    [Resource.DATA_SOURCES]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE],
    [Resource.EXPORTS]: [Action.CREATE, Action.READ, Action.DELETE, Action.EXPORT],
  },
  [Role.ADMIN]: {
    [Resource.DASHBOARD]: [Action.READ, Action.EXECUTE],
    [Resource.USERS]: [Action.READ, Action.UPDATE],
    [Resource.SETTINGS]: [Action.READ, Action.UPDATE],
    [Resource.REPORTS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.EXPORT],
    [Resource.ANALYTICS]: [Action.READ, Action.EXPORT],
    [Resource.AUDIT_LOGS]: [Action.READ],
    [Resource.NETWORK_ASSETS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.EXECUTE],
    [Resource.INTERVENTIONS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE, Action.APPROVE, Action.EXECUTE],
    [Resource.POLICIES]: [Action.READ, Action.UPDATE, Action.APPROVE],
    [Resource.ALERTS]: [Action.READ, Action.UPDATE],
    [Resource.DATA_SOURCES]: [Action.CREATE, Action.READ, Action.UPDATE, Action.DELETE],
    [Resource.EXPORTS]: [Action.CREATE, Action.READ, Action.DELETE, Action.EXPORT],
  },
  [Role.MANAGER]: {
    [Resource.DASHBOARD]: [Action.READ, Action.EXECUTE],
    [Resource.USERS]: [Action.READ],
    [Resource.SETTINGS]: [Action.READ],
    [Resource.REPORTS]: [Action.CREATE, Action.READ, Action.UPDATE, Action.EXPORT],
    [Resource.ANALYTICS]: [Action.READ, Action.EXPORT],
    [Resource.AUDIT_LOGS]: [Action.READ],
    [Resource.NETWORK_ASSETS]: [Action.READ, Action.UPDATE, Action.EXECUTE],
    [Resource.INTERVENTIONS]: [Action.READ, Action.UPDATE, Action.APPROVE],
    [Resource.POLICIES]: [Action.READ],
    [Resource.ALERTS]: [Action.READ, Action.UPDATE],
    [Resource.DATA_SOURCES]: [Action.READ],
    [Resource.EXPORTS]: [Action.CREATE, Action.READ, Action.EXPORT],
  },
  [Role.ANALYST]: {
    [Resource.DASHBOARD]: [Action.READ],
    [Resource.USERS]: [Action.READ],
    [Resource.SETTINGS]: [Action.READ],
    [Resource.REPORTS]: [Action.CREATE, Action.READ, Action.EXPORT],
    [Resource.ANALYTICS]: [Action.READ, Action.EXPORT],
    [Resource.AUDIT_LOGS]: [],
    [Resource.NETWORK_ASSETS]: [Action.READ],
    [Resource.INTERVENTIONS]: [Action.READ],
    [Resource.POLICIES]: [Action.READ],
    [Resource.ALERTS]: [Action.READ],
    [Resource.DATA_SOURCES]: [Action.READ],
    [Resource.EXPORTS]: [Action.READ, Action.EXPORT],
  },
  [Role.OPERATOR]: {
    [Resource.DASHBOARD]: [Action.READ, Action.EXECUTE],
    [Resource.USERS]: [],
    [Resource.SETTINGS]: [],
    [Resource.REPORTS]: [Action.READ],
    [Resource.ANALYTICS]: [Action.READ],
    [Resource.AUDIT_LOGS]: [],
    [Resource.NETWORK_ASSETS]: [Action.READ, Action.UPDATE],
    [Resource.INTERVENTIONS]: [Action.READ, Action.EXECUTE],
    [Resource.POLICIES]: [Action.READ],
    [Resource.ALERTS]: [Action.READ, Action.UPDATE],
    [Resource.DATA_SOURCES]: [Action.READ],
    [Resource.EXPORTS]: [],
  },
  [Role.VIEWER]: {
    [Resource.DASHBOARD]: [Action.READ],
    [Resource.USERS]: [],
    [Resource.SETTINGS]: [],
    [Resource.REPORTS]: [Action.READ],
    [Resource.ANALYTICS]: [Action.READ],
    [Resource.AUDIT_LOGS]: [],
    [Resource.NETWORK_ASSETS]: [Action.READ],
    [Resource.INTERVENTIONS]: [Action.READ],
    [Resource.POLICIES]: [Action.READ],
    [Resource.ALERTS]: [Action.READ],
    [Resource.DATA_SOURCES]: [Action.READ],
    [Resource.EXPORTS]: [],
  },
  [Role.GUEST]: {
    [Resource.DASHBOARD]: [],
    [Resource.USERS]: [],
    [Resource.SETTINGS]: [],
    [Resource.REPORTS]: [],
    [Resource.ANALYTICS]: [],
    [Resource.AUDIT_LOGS]: [],
    [Resource.NETWORK_ASSETS]: [],
    [Resource.INTERVENTIONS]: [],
    [Resource.POLICIES]: [],
    [Resource.ALERTS]: [],
    [Resource.DATA_SOURCES]: [],
    [Resource.EXPORTS]: [],
  },
};

/**
 * Routes that require authentication
 */
export const PROTECTED_ROUTES = [
  '/dashboard',
  '/analytics',
  '/reports',
  '/users',
  '/settings',
  '/network-assets',
  '/interventions',
  '/policies',
  '/audit-logs',
  '/data-sources',
];

/**
 * Routes accessible only to guests (not authenticated)
 */
export const GUEST_ROUTES = [
  '/login',
  '/register',
  '/forgot-password',
  '/reset-password',
];

/**
 * Routes accessible to specific roles only
 */
export const ROLE_PROTECTED_ROUTES: Record<string, Role[]> = {
  '/users': [Role.SUPER_ADMIN, Role.ADMIN],
  '/settings': [Role.SUPER_ADMIN, Role.ADMIN, Role.MANAGER],
  '/audit-logs': [Role.SUPER_ADMIN, Role.ADMIN, Role.MANAGER],
  '/policies': [Role.SUPER_ADMIN, Role.ADMIN, Role.MANAGER, Role.ANALYST],
  '/data-sources': [Role.SUPER_ADMIN, Role.ADMIN, Role.OPERATOR],
};

/**
 * Token configuration
 */
export const TOKEN_CONFIG = {
  STORAGE_KEY: 'neam_auth_token',
  REFRESH_STORAGE_KEY: 'neam_refresh_token',
  TOKEN_EXPIRY_BUFFER: 60 * 1000, // Refresh token 1 minute before expiry
  SESSION_TIMEOUT: 30 * 60 * 1000, // 30 minutes of inactivity
  REMEMBER_ME_EXPIRY: 7 * 24 * 60 * 60 * 1000, // 7 days
};

/**
 * Session configuration
 */
export const SESSION_CONFIG = {
  WARNING_BEFORE_EXPIRY: 2 * 60 * 1000, // Show warning 2 minutes before expiry
  IDLE_TIMEOUT: 15 * 60 * 1000, // 15 minutes of inactivity
  ABSOLUTE_TIMEOUT: 8 * 60 * 60 * 1000, // 8 hours absolute session limit
};
