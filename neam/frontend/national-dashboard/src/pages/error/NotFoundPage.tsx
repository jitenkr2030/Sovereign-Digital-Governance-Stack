import React from 'react';
import { Link } from 'react-router-dom';
import { Button } from '../../components/ui/Button';

/**
 * NotFoundPage Component
 *
 * 404 error page displayed when a user navigates to a non-existent route.
 */
export const NotFoundPage: React.FC = () => {
  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          {/* 404 Text */}
          <h1 className="text-9xl font-extrabold text-primary-600">404</h1>

          <h2 className="mt-6 text-3xl font-extrabold text-gray-900">
            Page not found
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            Sorry, we couldn't find the page you're looking for.
          </p>
        </div>

        <div className="mt-8 space-y-4">
          <Link to="/" className="block">
            <Button variant="primary" fullWidth>
              Go to Homepage
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

export default NotFoundPage;
