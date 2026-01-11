import React from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAppSelector } from '../../app/hooks';
import { selectIsAuthenticated } from '../../features/auth/authSlice';
import { LoadingSpinner } from '../../components/ui/LoadingSpinner';

/**
 * GuestGuard Component
 *
 * Protects routes that should only be accessible to guests (unauthenticated users).
 * Redirects authenticated users to the dashboard or specified redirect path.
 */
interface GuestGuardProps {
  children?: React.ReactNode;
  redirectPath?: string;
}

export const GuestGuard: React.FC<GuestGuardProps> = ({
  children,
  redirectPath = '/dashboard',
}) => {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const location = useLocation();

  // Determine where to redirect authenticated users
  // If they came from a specific page, try to redirect back there
  const from = location.state?.from || redirectPath;
  const shouldRedirect = isAuthenticated;

  if (shouldRedirect) {
    return <Navigate to={from} state={{ from: location.pathname }} replace />;
  }

  // User is not authenticated, render children or outlet
  return children ? <>{children}</> : <Outlet />;
};

/**
 * OnboardingGuard Component
 *
 * Special guard for onboarding flow.
 * Redirects users who have completed onboarding to the dashboard.
 * Redirects unauthenticated users to login.
 */
interface OnboardingGuardProps {
  children?: React.ReactNode;
  hasCompletedOnboarding: boolean;
  redirectPath?: string;
}

export const OnboardingGuard: React.FC<OnboardingGuardProps> = ({
  children,
  hasCompletedOnboarding,
  redirectPath = '/dashboard',
}) => {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />;
  }

  if (hasCompletedOnboarding) {
    return <Navigate to={redirectPath} replace />;
  }

  return children ? <>{children}</> : <Outlet />;
};

export default GuestGuard;
