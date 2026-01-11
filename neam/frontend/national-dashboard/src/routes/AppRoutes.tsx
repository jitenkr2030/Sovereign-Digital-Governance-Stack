import React from 'react';
import { Navigate, useRoutes, useLocation } from 'react-router-dom';
import { AuthGuard } from '../guards/AuthGuard';
import { RoleGuard } from '../guards/RoleGuard';
import { GuestGuard } from '../guards/GuestGuard';
import { ROLES } from '../features/auth/constants';
import { Layout } from '../components/layout/Layout';
import { LoadingSpinner } from '../components/ui/LoadingSpinner';

// Lazy load pages for code splitting
const LoginPage = React.lazy(
  () => import('../pages/auth/LoginPage').then((module) => ({ default: module.LoginPage }))
);
const RegisterPage = React.lazy(
  () => import('../pages/auth/RegisterPage').then((module) => ({ default: module.RegisterPage }))
);
const ForgotPasswordPage = React.lazy(
  () => import('../pages/auth/ForgotPasswordPage').then((module) => ({ default: module.ForgotPasswordPage }))
);
const DashboardPage = React.lazy(
  () => import('../pages/dashboard/DashboardPage').then((module) => ({ default: module.DashboardPage }))
);
const ProfilePage = React.lazy(
  () => import('../pages/profile/ProfilePage').then((module) => ({ default: module.ProfilePage }))
);
const SettingsPage = React.lazy(
  () => import('../pages/settings/SettingsPage').then((module) => ({ default: module.SettingsPage }))
);
const UsersPage = React.lazy(
  () => import('../pages/admin/UsersPage').then((module) => ({ default: module.UsersPage }))
);
const AccessDeniedPage = React.lazy(
  () => import('../pages/error/AccessDeniedPage').then((module) => ({ default: module.AccessDeniedPage }))
);
const NotFoundPage = React.lazy(
  () => import('../pages/error/NotFoundPage').then((module) => ({ default: module.NotFoundPage }))
);
const OnboardingPage = React.lazy(
  () => import('../pages/onboarding/OnboardingPage').then((module) => ({ default: module.OnboardingPage }))
);

// Loading fallback component
const PageLoader: React.FC<{ children?: React.ReactNode }> = ({ children }) => (
  <div className="flex items-center justify-center min-h-screen">
    <LoadingSpinner size="lg" label="Loading page..." />
    {children}
  </div>
);

// Suspense wrapper for lazy loaded pages
const SuspensedPage: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <React.Suspense fallback={<PageLoader />}>{children}</React.Suspense>
);

/**
 * AppRoutes Configuration
 *
 * Central routing configuration for the application.
 * Defines all routes with their corresponding guards and page components.
 */
export const AppRoutes: React.FC = () => {
  const location = useLocation();

  // Define all routes in the application
  const routes = [
    // Public routes (accessible to everyone)
    {
      path: '/',
      element: <Layout />,
      children: [
        {
          index: true,
          element: (
            <SuspensedPage>
              <Navigate to="/dashboard" replace />
            </SuspensedPage>
          ),
        },
      ],
    },
    // Guest-only routes (accessible to unauthenticated users)
    {
      path: '/auth',
      element: (
        <GuestGuard>
          <Layout variant="auth" />
        </GuestGuard>
      ),
      children: [
        {
          path: 'login',
          element: (
            <SuspensedPage>
              <LoginPage />
            </SuspensedPage>
          ),
        },
        {
          path: 'register',
          element: (
            <SuspensedPage>
              <RegisterPage />
            </SuspensedPage>
          ),
        },
        {
          path: 'forgot-password',
          element: (
            <SuspensedPage>
              <ForgotPasswordPage />
            </SuspensedPage>
          ),
        },
      ],
    },
    // Protected routes (require authentication)
    {
      path: '/',
      element: (
        <AuthGuard>
          <Layout />
        </AuthGuard>
      ),
      children: [
        {
          path: 'dashboard',
          element: (
            <SuspensedPage>
              <DashboardPage />
            </SuspensedPage>
          ),
        },
        {
          path: 'profile',
          element: (
            <SuspensedPage>
              <ProfilePage />
            </SuspensedPage>
          ),
        },
        {
          path: 'settings',
          element: (
            <SuspensedPage>
              <SettingsPage />
            </SuspensedPage>
          ),
        },
        // Onboarding route (temporary, redirects if already completed)
        {
          path: 'onboarding',
          element: (
            <SuspensedPage>
              <OnboardingPage />
            </SuspensedPage>
          ),
        },
      ],
    },
    // Admin-only routes (require ADMIN or SUPER_ADMIN role)
    {
      path: '/admin',
      element: (
        <AuthGuard>
          <RoleGuard
            allowedRoles={[ROLES.ADMIN, ROLES.SUPER_ADMIN]}
            fallbackPath="/access-denied"
          >
            <Layout />
          </RoleGuard>
        </AuthGuard>
      ),
      children: [
        {
          path: 'users',
          element: (
            <SuspensedPage>
              <UsersPage />
            </SuspensedPage>
          ),
        },
      ],
    },
    // Super admin-only routes (require SUPER_ADMIN role)
    {
      path: '/super-admin',
      element: (
        <AuthGuard>
          <RoleGuard
            allowedRoles={[ROLES.SUPER_ADMIN]}
            fallbackPath="/access-denied"
          >
            <Layout variant="admin" />
          </RoleGuard>
        </AuthGuard>
      ),
      children: [
        // Add super admin routes here
      ],
    },
    // Error pages
    {
      path: '/access-denied',
      element: (
        <SuspensedPage>
          <AccessDeniedPage />
        </SuspensedPage>
      ),
    },
    // 404 catch-all route
    {
      path: '*',
      element: (
        <SuspensedPage>
          <NotFoundPage />
        </SuspensedPage>
      ),
    },
  ];

  // Use react-router's useRoutes hook
  const element = useRoutes(routes);

  return (
    <React.Fragment>
      {/* Global route change listener for analytics, scroll management, etc. */}
      <RouteChangeListener pathname={location.pathname} />
      {element}
    </React.Fragment>
  );
};

/**
 * RouteChangeListener Component
 *
 * Handles side effects on route changes.
 * Useful for analytics tracking, scroll restoration, etc.
 */
const RouteChangeListener: React.FC<{ pathname: string }> = ({ pathname }) => {
  React.useEffect(() => {
    // Scroll to top on route change
    window.scrollTo(0, 0);

    // You can add analytics tracking here
    // analytics.trackPageView(pathname);
    console.log(`Route changed to: ${pathname}`);
  }, [pathname]);

  return null;
};

/**
 * getProtectedRoutePath helper function
 *
 * Returns the appropriate redirect path based on user role.
 * Useful for login redirects.
 */
export const getProtectedRoutePath = (role: string): string => {
  switch (role) {
    case ROLES.SUPER_ADMIN:
      return '/super-admin';
    case ROLES.ADMIN:
      return '/admin';
    default:
      return '/dashboard';
  }
};

export default AppRoutes;
