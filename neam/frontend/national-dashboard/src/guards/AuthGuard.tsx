import React, { useEffect, useState } from 'react';
import { Navigate, Outlet, useLocation } from 'react-router-dom';
import { useAppSelector } from '../../app/hooks';
import {
  selectIsAuthenticated,
  selectIsSessionExpired,
  selectAuthLoading,
} from '../../features/auth/authSlice';
import { LoadingSpinner } from '../../components/ui/LoadingSpinner';

/**
 * AuthGuard Component
 *
 * Protects routes that require authentication.
 * Redirects unauthenticated users to the login page.
 * Redirects authenticated users away from login/register pages.
 */
export const AuthGuard: React.FC<{
  children?: React.ReactNode;
  fallbackPath?: string;
}> = ({ children, fallbackPath = '/login' }) => {
  const isAuthenticated = useAppSelector(selectIsAuthenticated);
  const isSessionExpired = useAppSelector(selectIsSessionExpired);
  const isLoading = useAppSelector(selectAuthLoading);
  const location = useLocation();
  const [hasCheckedSession, setHasCheckedSession] = useState(false);

  useEffect(() => {
    // Wait for initial session check to complete
    if (!isLoading) {
      setHasCheckedSession(true);
    }
  }, [isLoading]);

  // Show loading spinner while checking authentication status
  if (!hasCheckedSession) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner size="lg" label="Verifying session..." />
      </div>
    );
  }

  // If user is not authenticated, redirect to login
  if (!isAuthenticated || isSessionExpired) {
    return (
      <Navigate
        to={fallbackPath}
        state={{ from: location.pathname }}
        replace
      />
    );
  }

  // If user is authenticated, render children or outlet
  return children ? <>{children}</> : <Outlet />;
};

export default AuthGuard;
