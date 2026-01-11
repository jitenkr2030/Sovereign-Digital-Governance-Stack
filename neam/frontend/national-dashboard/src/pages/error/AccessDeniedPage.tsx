import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Button } from '../../components/ui/Button';

/**
 * AccessDeniedPage Component
 *
 * Displayed when a user attempts to access a route they don't have permission for.
 */
export const AccessDeniedPage: React.FC = () => {
  const location = useLocation();
  const from = location.state?.from || '/dashboard';

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          {/* Error Icon */}
          <div className="mx-auto flex items-center justify-center h-16 w-16 rounded-full bg-red-100">
            <svg
              className="h-10 w-10 text-red-600"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"
              />
            </svg>
          </div>

          <h2 className="mt-6 text-3xl font-extrabold text-gray-900">
            Access Denied
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            You don't have permission to access this page. If you believe this
            is an error, please contact your administrator.
          </p>
        </div>

        <div className="mt-8 space-y-4">
          <Link to={from} className="block">
            <Button variant="primary" fullWidth>
              Return to {from === '/dashboard' ? 'Dashboard' : 'Previous Page'}
            </Button>
          </Link>
          <Link to="/dashboard" className="block">
            <Button variant="outline" fullWidth>
              Go to Dashboard
            </Button>
          </Link>
        </div>
      </div>
    </div>
  );
};

export default AccessDeniedPage;
