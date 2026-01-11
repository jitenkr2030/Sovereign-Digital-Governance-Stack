import React from 'react';
import { useAppSelector } from '../../app/hooks';
import { selectCurrentUser } from '../../features/auth/authSlice';
import { Button } from '../../components/ui/Button';
import { Card } from '../../components/ui/Card';
import { Input } from '../../components/ui/Input';
import { LoadingSpinner } from '../../components/ui/LoadingSpinner';

/**
 * ProfilePage Component
 *
 * User profile page displaying and allowing editing of user information.
 * Protected route requiring authentication.
 */
export const ProfilePage: React.FC = () => {
  const user = useAppSelector(selectCurrentUser);

  if (!user) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" label="Loading profile..." />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Profile</h1>
        <p className="mt-1 text-sm text-gray-500">
          Manage your personal information and account settings.
        </p>
      </div>

      {/* Profile Information */}
      <Card>
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">
            Personal Information
          </h2>
        </div>
        <div className="p-6">
          <div className="flex items-center space-x-6 mb-6">
            {/* Avatar */}
            <div className="flex-shrink-0">
              <div className="h-20 w-20 rounded-full bg-primary-100 flex items-center justify-center">
                <span className="text-2xl font-bold text-primary-600">
                  {user.firstName[0]}
                  {user.lastName[0]}
                </span>
              </div>
            </div>
            <div>
              <Button variant="outline" size="sm">
                Change avatar
              </Button>
              <p className="mt-1 text-xs text-gray-500">
                JPG, GIF or PNG. Max size 1MB.
              </p>
            </div>
          </div>

          {/* Name Fields */}
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label
                htmlFor="firstName"
                className="block text-sm font-medium text-gray-700"
              >
                First name
              </label>
              <Input
                id="firstName"
                name="firstName"
                type="text"
                defaultValue={user.firstName}
                className="mt-1"
              />
            </div>
            <div>
              <label
                htmlFor="lastName"
                className="block text-sm font-medium text-gray-700"
              >
                Last name
              </label>
              <Input
                id="lastName"
                name="lastName"
                type="text"
                defaultValue={user.lastName}
                className="mt-1"
              />
            </div>
          </div>

          {/* Email Field */}
          <div className="mt-6">
            <label
              htmlFor="email"
              className="block text-sm font-medium text-gray-700"
            >
              Email address
            </label>
            <Input
              id="email"
              name="email"
              type="email"
              defaultValue={user.email}
              className="mt-1"
            />
          </div>

          {/* Save Button */}
          <div className="mt-6">
            <Button variant="primary">Save changes</Button>
          </div>
        </div>
      </Card>

      {/* Account Security */}
      <Card>
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-medium text-gray-900">Account Security</h2>
        </div>
        <div className="p-6 space-y-6">
          {/* Change Password */}
          <div>
            <h3 className="text-sm font-medium text-gray-900">Password</h3>
            <p className="mt-1 text-sm text-gray-500">
              Keep your account secure by using a strong password.
            </p>
            <div className="mt-3">
              <Button variant="outline">Change password</Button>
            </div>
          </div>

          {/* Two-Factor Authentication */}
          <div className="pt-6 border-t border-gray-200">
            <h3 className="text-sm font-medium text-gray-900">
              Two-factor authentication
            </h3>
            <p className="mt-1 text-sm text-gray-500">
              Add an extra layer of security to your account.
            </p>
            <div className="mt-3">
              <Button variant="outline">Enable 2FA</Button>
            </div>
          </div>
        </div>
      </Card>

      {/* Danger Zone */}
      <Card className="border-red-200">
        <div className="px-6 py-4 border-b border-red-200 bg-red-50">
          <h2 className="text-lg font-medium text-red-900">Danger Zone</h2>
        </div>
        <div className="p-6">
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-sm font-medium text-gray-900">
                Delete account
              </h3>
              <p className="mt-1 text-sm text-gray-500">
                Permanently delete your account and all associated data.
              </p>
            </div>
            <Button variant="danger">Delete account</Button>
          </div>
        </div>
      </Card>
    </div>
  );
};

export default ProfilePage;
