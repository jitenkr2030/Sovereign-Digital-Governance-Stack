import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useTheme } from '../context/ThemeContext';
import { Button, Card, CardHeader, CardTitle, CardBody } from '../components/common';
import {
  User,
  Bell,
  Shield,
  Palette,
  Database,
  Key,
  Globe,
  Moon,
  Sun,
  Monitor,
  Save,
  RefreshCw,
} from 'lucide-react';

export default function Settings() {
  const { user } = useAuth();
  const { theme, isDark, toggleTheme, setThemeMode } = useTheme();
  const [activeTab, setActiveTab] = useState('profile');
  const [isSaving, setIsSaving] = useState(false);

  // Profile settings
  const [profile, setProfile] = useState({
    name: user?.name || '',
    email: user?.email || '',
    role: user?.role || 'regulator',
  });

  // Notification settings
  const [notifications, setNotifications] = useState({
    email: true,
    push: true,
    licenseAlerts: true,
    energyAlerts: true,
    reportAlerts: true,
    weeklyDigest: false,
  });

  // Security settings
  const [security, setSecurity] = useState({
    twoFactor: false,
    sessionTimeout: 30,
    ipWhitelist: '',
  });

  // Integration settings
  const [integrations, setIntegrations] = useState({
    licensingService: 'http://localhost:3000',
    energyService: 'http://localhost:8000',
    reportingService: 'http://localhost:3001',
  });

  const handleSave = async () => {
    setIsSaving(true);
    // Simulate API call
    await new Promise((resolve) => setTimeout(resolve, 1000));
    setIsSaving(false);
  };

  const tabs = [
    { id: 'profile', label: 'Profile', icon: User },
    { id: 'notifications', label: 'Notifications', icon: Bell },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'appearance', label: 'Appearance', icon: Palette },
    { id: 'integrations', label: 'Integrations', icon: Globe },
    { id: 'api', label: 'API Keys', icon: Key },
  ];

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-50">Settings</h1>
        <p className="text-dark-500 dark:text-dark-400 mt-1">
          Manage your account settings and preferences
        </p>
      </div>

      <div className="flex flex-col lg:flex-row gap-6">
        {/* Sidebar */}
        <div className="lg:w-64 flex-shrink-0">
          <Card padding="none">
            <nav className="p-2">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-left transition-colors ${
                    activeTab === tab.id
                      ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-700 dark:text-primary-300'
                      : 'text-dark-600 dark:text-dark-400 hover:bg-dark-100 dark:hover:bg-dark-800'
                  }`}
                >
                  <tab.icon className="w-5 h-5" />
                  <span className="font-medium">{tab.label}</span>
                </button>
              ))}
            </nav>
          </Card>
        </div>

        {/* Content */}
        <div className="flex-1">
          {/* Profile Settings */}
          {activeTab === 'profile' && (
            <Card>
              <CardHeader>
                <CardTitle>Profile Settings</CardTitle>
              </CardHeader>
              <CardBody className="space-y-6">
                <div className="flex items-center gap-6">
                  <div className="w-20 h-20 rounded-full bg-primary-600 flex items-center justify-center text-white text-2xl font-bold">
                    {profile.name.charAt(0).toUpperCase()}
                  </div>
                  <div>
                    <Button variant="secondary" size="sm">Change Avatar</Button>
                    <p className="text-sm text-dark-500 mt-2">JPG, GIF or PNG. Max 2MB.</p>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label className="label">Full Name</label>
                    <input
                      type="text"
                      value={profile.name}
                      onChange={(e) => setProfile({ ...profile, name: e.target.value })}
                      className="input"
                    />
                  </div>
                  <div>
                    <label className="label">Email Address</label>
                    <input
                      type="email"
                      value={profile.email}
                      onChange={(e) => setProfile({ ...profile, email: e.target.value })}
                      className="input"
                    />
                  </div>
                  <div>
                    <label className="label">Role</label>
                    <select
                      value={profile.role}
                      onChange={(e) => setProfile({ ...profile, role: e.target.value })}
                      className="input"
                    >
                      <option value="admin">Administrator</option>
                      <option value="regulator">Regulator</option>
                      <option value="auditor">Auditor</option>
                      <option value="viewer">Viewer</option>
                    </select>
                  </div>
                </div>

                <div className="flex justify-end pt-4 border-t">
                  <Button onClick={handleSave} isLoading={isSaving} leftIcon={<Save className="w-4 h-4" />}>
                    Save Changes
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

          {/* Notification Settings */}
          {activeTab === 'notifications' && (
            <Card>
              <CardHeader>
                <CardTitle>Notification Preferences</CardTitle>
              </CardHeader>
              <CardBody className="space-y-6">
                <div className="space-y-4">
                  <h4 className="font-medium">Delivery Methods</h4>
                  {[
                    { key: 'email', label: 'Email Notifications', description: 'Receive notifications via email' },
                    { key: 'push', label: 'Push Notifications', description: 'Browser push notifications' },
                  ].map((item) => (
                    <div key={item.key} className="flex items-center justify-between">
                      <div>
                        <p className="font-medium">{item.label}</p>
                        <p className="text-sm text-dark-500">{item.description}</p>
                      </div>
                      <button
                        onClick={() =>
                          setNotifications({ ...notifications, [item.key]: !notifications[item.key as keyof typeof notifications] })
                        }
                        className={`relative w-12 h-6 rounded-full transition-colors ${
                          notifications[item.key as keyof typeof notifications] ? 'bg-primary-600' : 'bg-dark-300 dark:bg-dark-600'
                        }`}
                      >
                        <span
                          className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${
                            notifications[item.key as keyof typeof notifications] ? 'translate-x-6' : ''
                          }`}
                        />
                      </button>
                    </div>
                  ))}
                </div>

                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Alert Types</h4>
                  {[
                    { key: 'licenseAlerts', label: 'License Alerts', description: 'License status changes and renewals' },
                    { key: 'energyAlerts', label: 'Energy Alerts', description: 'Grid status and consumption alerts' },
                    { key: 'reportAlerts', label: 'Report Alerts', description: 'Report generation status' },
                    { key: 'weeklyDigest', label: 'Weekly Digest', description: 'Summary of weekly activity' },
                  ].map((item) => (
                    <div key={item.key} className="flex items-center justify-between">
                      <div>
                        <p className="font-medium">{item.label}</p>
                        <p className="text-sm text-dark-500">{item.description}</p>
                      </div>
                      <button
                        onClick={() =>
                          setNotifications({ ...notifications, [item.key]: !notifications[item.key as keyof typeof notifications] })
                        }
                        className={`relative w-12 h-6 rounded-full transition-colors ${
                          notifications[item.key as keyof typeof notifications] ? 'bg-primary-600' : 'bg-dark-300 dark:bg-dark-600'
                        }`}
                      >
                        <span
                          className={`absolute top-1 left-1 w-4 h-4 bg-white rounded-full transition-transform ${
                            notifications[item.key as keyof typeof notifications] ? 'translate-x-6' : ''
                          }`}
                        />
                      </button>
                    </div>
                  ))}
                </div>

                <div className="flex justify-end pt-4 border-t">
                  <Button onClick={handleSave} isLoading={isSaving} leftIcon={<Save className="w-4 h-4" />}>
                    Save Preferences
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

          {/* Security Settings */}
          {activeTab === 'security' && (
            <Card>
              <CardHeader>
                <CardTitle>Security Settings</CardTitle>
              </CardHeader>
              <CardBody className="space-y-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="font-medium">Two-Factor Authentication</p>
                    <p className="text-sm text-dark-500">Add an extra layer of security to your account</p>
                  </div>
                  <Button variant={security.twoFactor ? 'danger' : 'secondary'}>
                    {security.twoFactor ? 'Disable 2FA' : 'Enable 2FA'}
                  </Button>
                </div>

                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Session Settings</h4>
                  <div>
                    <label className="label">Session Timeout (minutes)</label>
                    <select
                      value={security.sessionTimeout}
                      onChange={(e) => setSecurity({ ...security, sessionTimeout: Number(e.target.value) })}
                      className="input max-w-xs"
                    >
                      <option value={15}>15 minutes</option>
                      <option value={30}>30 minutes</option>
                      <option value={60}>1 hour</option>
                      <option value={120}>2 hours</option>
                    </select>
                  </div>
                </div>

                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Password</h4>
                  <Button variant="secondary" className="w-auto">
                    Change Password
                  </Button>
                </div>

                <div className="space-y-4 pt-4 border-t">
                  <h4 className="font-medium">Active Sessions</h4>
                  <div className="space-y-2">
                    {[
                      { device: 'Chrome on Windows', location: 'New York, US', current: true },
                      { device: 'Safari on macOS', location: 'London, UK', current: false },
                    ].map((session, index) => (
                      <div
                        key={index}
                        className="flex items-center justify-between p-3 rounded-lg border border-dark-200 dark:border-dark-700"
                      >
                        <div>
                          <p className="font-medium">{session.device}</p>
                          <p className="text-sm text-dark-500">{session.location}</p>
                        </div>
                        {session.current ? (
                          <span className="badge badge-success">Current</span>
                        ) : (
                          <Button variant="ghost" size="sm">
                            Revoke
                          </Button>
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                <div className="flex justify-end pt-4 border-t">
                  <Button onClick={handleSave} isLoading={isSaving} leftIcon={<Save className="w-4 h-4" />}>
                    Save Settings
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

          {/* Appearance Settings */}
          {activeTab === 'appearance' && (
            <Card>
              <CardHeader>
                <CardTitle>Appearance</CardTitle>
              </CardHeader>
              <CardBody className="space-y-6">
                <div>
                  <h4 className="font-medium mb-4">Theme</h4>
                  <div className="grid grid-cols-3 gap-4">
                    {[
                      { value: 'light', label: 'Light', icon: Sun },
                      { value: 'dark', label: 'Dark', icon: Moon },
                      { value: 'system', label: 'System', icon: Monitor },
                    ].map((option) => (
                      <button
                        key={option.value}
                        onClick={() => setThemeMode(option.value as 'light' | 'dark' | 'system')}
                        className={`p-4 rounded-lg border-2 transition-colors ${
                          theme.mode === option.value
                            ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                            : 'border-dark-200 dark:border-dark-700 hover:border-dark-300'
                        }`}
                      >
                        <option.icon className="w-6 h-6 mx-auto mb-2" />
                        <p className="font-medium">{option.label}</p>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="flex justify-end pt-4 border-t">
                  <Button onClick={handleSave} isLoading={isSaving} leftIcon={<Save className="w-4 h-4" />}>
                    Save Preferences
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

          {/* Integrations Settings */}
          {activeTab === 'integrations' && (
            <Card>
              <CardHeader>
                <CardTitle>Service Integrations</CardTitle>
              </CardHeader>
              <CardBody className="space-y-6">
                <div className="space-y-4">
                  {[
                    { key: 'licensingService', label: 'Licensing Service', description: 'Node.js/NestJS service for VASP/CASP licensing' },
                    { key: 'energyService', label: 'Energy Service', description: 'Python/FastAPI service for energy analytics' },
                    { key: 'reportingService', label: 'Reporting Service', description: 'Node.js/Express service for report generation' },
                  ].map((service) => (
                    <div key={service.key}>
                      <label className="label">{service.label}</label>
                      <p className="text-sm text-dark-500 mb-2">{service.description}</p>
                      <input
                        type="url"
                        value={integrations[service.key as keyof typeof integrations]}
                        onChange={(e) =>
                          setIntegrations({ ...integrations, [service.key]: e.target.value })
                        }
                        className="input"
                        placeholder="http://localhost:3000"
                      />
                    </div>
                  ))}
                </div>

                <div className="flex justify-end gap-3 pt-4 border-t">
                  <Button variant="secondary" leftIcon={<RefreshCw className="w-4 h-4" />}>
                    Test Connections
                  </Button>
                  <Button onClick={handleSave} isLoading={isSaving} leftIcon={<Save className="w-4 h-4" />}>
                    Save Settings
                  </Button>
                </div>
              </CardBody>
            </Card>
          )}

          {/* API Keys Settings */}
          {activeTab === 'api' && (
            <Card>
              <CardHeader>
                <CardTitle>API Keys</CardTitle>
                <Button size="sm" leftIcon={<Key className="w-4 h-4" />}>
                  Generate New Key
                </Button>
              </CardHeader>
              <CardBody className="space-y-4">
                <div className="p-4 rounded-lg border border-dark-200 dark:border-dark-700">
                  <div className="flex items-center justify-between mb-2">
                    <p className="font-medium">Production API Key</p>
                    <span className="badge badge-success">Active</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-2 bg-dark-100 dark:bg-dark-800 rounded font-mono text-sm">
                      csic_live_•••••••••••••••••••••••••
                    </code>
                    <Button variant="ghost" size="sm">Copy</Button>
                  </div>
                  <p className="text-sm text-dark-500 mt-2">Created: Jan 1, 2024</p>
                </div>

                <div className="p-4 rounded-lg border border-dark-200 dark:border-dark-700 opacity-60">
                  <div className="flex items-center justify-between mb-2">
                    <p className="font-medium">Development API Key</p>
                    <span className="badge badge-neutral">Revoked</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <code className="flex-1 p-2 bg-dark-100 dark:bg-dark-800 rounded font-mono text-sm">
                      csic_dev_•••••••••••••••••••••••••
                    </code>
                  </div>
                  <p className="text-sm text-dark-500 mt-2">Created: Dec 15, 2023</p>
                </div>
              </CardBody>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}
