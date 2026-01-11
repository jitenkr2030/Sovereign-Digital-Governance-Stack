import React from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { useAppSelector } from '../../app/hooks';
import { selectCurrentUser } from '../../features/auth/authSlice';
import { ROLES } from '../../features/auth/constants';
import type { Role, Permission } from '../../features/auth/types';
import { LoadingSpinner } from '../../components/ui/LoadingSpinner';

/**
 * RoleGuard Component
 *
 * Protects routes based on user roles.
 * Redirects users without the required role to the access denied page or dashboard.
 */
interface RoleGuardProps {
  children?: React.ReactNode;
  allowedRoles: Role[];
  requiredPermissions?: Permission[];
  requireAllPermissions?: boolean;
  fallbackPath?: string;
  redirectOnNoRole?: boolean;
}

export const RoleGuard: React.FC<RoleGuardProps> = ({
  children,
  allowedRoles,
  requiredPermissions = [],
  requireAllPermissions = false,
  fallbackPath = '/access-denied',
  redirectOnNoRole = true,
}) => {
  const user = useAppSelector(selectCurrentUser);
  const location = useLocation();

  // If user is not loaded yet, show loading state
  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" label="Loading user permissions..." />
      </div>
    );
  }

  // Check if user's role is in the allowed roles list
  const hasAllowedRole = allowedRoles.includes(user.role as Role);

  // Check if user has required permissions
  const userPermissions = user.permissions || [];
  const hasRequiredPermissions = requiredPermissions.length > 0
    ? requireAllPermissions
      ? requiredPermissions.every((perm) => userPermissions.includes(perm))
      : requiredPermissions.some((perm) => userPermissions.includes(perm))
    : true;

  // Determine access status
  const hasAccess = hasAllowedRole && hasRequiredPermissions;

  if (!hasAccess && redirectOnNoRole) {
    // Store the attempted URL for redirecting after login
    return (
      <Navigate
        to={fallbackPath}
        state={{ from: location.pathname, reason: 'insufficient_permissions' }}
        replace
      />
    );
  }

  if (!hasAccess) {
    // Return null or a restricted content component instead of redirecting
    return null;
  }

  // User has access, render children
  return children ? <>{children}</> : null;
};

/**
 * PermissionGuard Component (simplified wrapper for single permission checks)
 */
interface PermissionGuardProps {
  children?: React.ReactNode;
  permission: Permission;
  fallback?: React.ReactNode;
}

export const PermissionGuard: React.FC<PermissionGuardProps> = ({
  children,
  permission,
  fallback = null,
}) => {
  const user = useAppSelector(selectCurrentUser);

  if (!user) {
    return <>{fallback}</>;
  }

  const hasPermission =
    user.role === ROLES.ADMIN ||
    (user.permissions || []).includes(permission);

  return hasPermission ? <>{children}</> : <>{fallback}</>;
};

/**
 * AdminGuard Component
 *专门用于管理员访问的路由守卫
 */
export const AdminGuard: React.FC<{ children?: React.ReactNode }> = ({
  children,
}) => {
  return (
    <RoleGuard
      allowedRoles={[ROLES.ADMIN]}
      fallbackPath="/dashboard"
    >
      {children}
    </RoleGuard>
  );
};

/**
 * SuperAdminGuard Component
 *专门用于超级管理员访问的路由守卫
 */
export const SuperAdminGuard: React.FC<{ children?: React.ReactNode }> = ({
  children,
}) => {
  return (
    <RoleGuard
      allowedRoles={[ROLES.SUPER_ADMIN]}
      fallbackPath="/access-denied"
    >
      {children}
    </RoleGuard>
  );
};

export default RoleGuard;
