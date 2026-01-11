import React from 'react';
import { useAppSelector } from '../../app/hooks';
import { selectCurrentUser } from '../../features/auth/authSlice';
import { ROLES } from '../../features/auth/constants';
import { Card } from '../../components/ui/Card';
import { LoadingSpinner } from '../../components/ui/LoadingSpinner';

/**
 * DashboardPage Component
 *
 * Main dashboard page displaying key metrics and overview information.
 * Protected route requiring authentication.
 */
export const DashboardPage: React.FC = () => {
  const user = useAppSelector(selectCurrentUser);

  if (!user) {
    return (
      <div className="flex items-center justify-center h-full">
        <LoadingSpinner size="lg" label="Loading dashboard..." />
      </div>
    );
  }

  const greeting = getGreeting();
  const roleDisplayName = getRoleDisplayName(user.role);

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">
          {greeting}, {user.firstName}!
        </h1>
        <p className="mt-1 text-sm text-gray-500">
          Welcome back to your dashboard. Here's what's happening today.
        </p>
      </div>

      {/* Quick Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Total Users"
          value="1,234"
          change="+12%"
          changeType="increase"
          icon="users"
        />
        <StatCard
          title="Active Sessions"
          value="456"
          change="+8%"
          changeType="increase"
          icon="activity"
        />
        <StatCard
          title="Pending Tasks"
          value="23"
          change="-3%"
          changeType="decrease"
          icon="tasks"
        />
        <StatCard
          title="Revenue"
          value="$12,345"
          change="+15%"
          changeType="increase"
          icon="revenue"
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Recent Activity */}
        <Card className="lg:col-span-2">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">
              Recent Activity
            </h2>
          </div>
          <div className="p-6">
            <ActivityList />
          </div>
        </Card>

        {/* User Info Sidebar */}
        <Card>
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">
              Your Profile
            </h2>
          </div>
          <div className="p-6 space-y-4">
            <div>
              <p className="text-sm font-medium text-gray-500">Name</p>
              <p className="text-base text-gray-900">
                {user.firstName} {user.lastName}
              </p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Email</p>
              <p className="text-base text-gray-900">{user.email}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-gray-500">Role</p>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                {roleDisplayName}
              </span>
            </div>
          </div>
        </Card>
      </div>
    </div>
  );
};

/**
 * Helper function to get greeting based on time of day
 */
function getGreeting(): string {
  const hour = new Date().getHours();
  if (hour < 12) return 'Good morning';
  if (hour < 18) return 'Good afternoon';
  return 'Good evening';
}

/**
 * Helper function to get display name for role
 */
function getRoleDisplayName(role: string): string {
  const roleNames: Record<string, string> = {
    [ROLES.SUPER_ADMIN]: 'Super Admin',
    [ROLES.ADMIN]: 'Administrator',
    [ROLES.MANAGER]: 'Manager',
    [ROLES.USER]: 'User',
    [ROLES.GUEST]: 'Guest',
  };
  return roleNames[role] || role;
}

/**
 * StatCard Component
 */
interface StatCardProps {
  title: string;
  value: string;
  change: string;
  changeType: 'increase' | 'decrease';
  icon: string;
}

const StatCard: React.FC<StatCardProps> = ({
  title,
  value,
  change,
  changeType,
}) => {
  const changeColor =
    changeType === 'increase' ? 'text-green-600' : 'text-red-600';

  return (
    <Card>
      <div className="p-6">
        <p className="text-sm font-medium text-gray-500 truncate">{title}</p>
        <div className="mt-1 flex items-baseline">
          <p className="text-2xl font-semibold text-gray-900">{value}</p>
          <p className={`ml-2 text-sm font-medium ${changeColor}`}>{change}</p>
        </div>
      </div>
    </Card>
  );
};

/**
 * ActivityList Component
 */
const ActivityList: React.FC = () => {
  const activities = [
    {
      id: 1,
      action: 'User registration',
      user: 'John Doe',
      time: '5 minutes ago',
    },
    {
      id: 2,
      action: 'Report generated',
      user: 'Jane Smith',
      time: '15 minutes ago',
    },
    {
      id: 3,
      action: 'Settings updated',
      user: 'Bob Johnson',
      time: '1 hour ago',
    },
    {
      id: 4,
      action: 'New order placed',
      user: 'Alice Williams',
      time: '2 hours ago',
    },
  ];

  return (
    <ul className="divide-y divide-gray-200">
      {activities.map((activity) => (
        <li key={activity.id} className="py-4">
          <div className="flex space-x-3">
            <div className="flex-1 space-y-1">
              <div className="flex items-center justify-between">
                <h3 className="text-sm font-medium">{activity.action}</h3>
                <p className="text-sm text-gray-500">{activity.time}</p>
              </div>
              <p className="text-sm text-gray-500">
                by {activity.user}
              </p>
            </div>
          </div>
        </li>
      ))}
    </ul>
  );
};

export default DashboardPage;
