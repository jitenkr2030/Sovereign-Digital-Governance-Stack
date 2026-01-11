import React, { useState } from 'react';
import { Card, CardBody, CardHeader, CardTitle, Select, Input, Button, Table } from '../../components/common';

const Admin: React.FC = () => {
  const [activeTab, setActiveTab] = useState('users');

  const users = [
    { id: '1', name: 'John Doe', email: 'john@csic.gov', role: 'admin', status: 'active' },
    { id: '2', name: 'Jane Smith', email: 'jane@csic.gov', role: 'officer', status: 'active' },
    { id: '3', name: 'Bob Wilson', email: 'bob@csic.gov', role: 'auditor', status: 'inactive' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Administration</h1>
          <p className="text-slate-500 mt-1">System administration and user management</p>
        </div>
        <Button variant="primary">Add User</Button>
      </div>

      {/* Tabs */}
      <div className="border-b border-slate-200">
        <nav className="flex space-x-8">
          {['users', 'roles', 'settings', 'integrations'].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab)}
              className={`
                py-4 px-1 border-b-2 font-medium text-sm capitalize transition-colors
                ${activeTab === tab
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-slate-500 hover:text-slate-700'
                }
              `}
            >
              {tab}
            </button>
          ))}
        </nav>
      </div>

      {/* Users Tab */}
      {activeTab === 'users' && (
        <Card>
          <CardBody>
            <div className="mb-4 flex items-center gap-4">
              <Input placeholder="Search users..." fullWidth className="max-w-xs" />
              <Select
                value=""
                onChange={() => {}}
                options={[
                  { value: '', label: 'All Roles' },
                  { value: 'admin', label: 'Admin' },
                  { value: 'officer', label: 'Officer' },
                  { value: 'auditor', label: 'Auditor' },
                ]}
              />
            </div>
            <Table
              headers={['Name', 'Email', 'Role', 'Status', 'Actions']}
              data={users.map((user) => ({
                ...user,
                actions: (
                  <div className="flex gap-2">
                    <Button variant="secondary" size="sm">Edit</Button>
                    <Button variant="danger" size="sm">Delete</Button>
                  </div>
                ),
              }))}
            />
          </CardBody>
        </Card>
      )}

      {/* Roles Tab */}
      {activeTab === 'roles' && (
        <Card>
          <CardBody>
            <div className="space-y-4">
              {['admin', 'officer', 'auditor', 'viewer'].map((role) => (
                <div key={role} className="flex items-center justify-between p-4 bg-slate-50 rounded-lg">
                  <div>
                    <p className="font-medium text-slate-900 capitalize">{role}</p>
                    <p className="text-sm text-slate-500">Full access to all features</p>
                  </div>
                  <Button variant="secondary" size="sm">Configure Permissions</Button>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>
      )}

      {/* Settings Tab */}
      {activeTab === 'settings' && (
        <Card>
          <CardHeader>
            <CardTitle>System Settings</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="space-y-6">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Session Timeout (minutes)
                </label>
                <Input type="number" defaultValue={30} className="max-w-xs" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-2">
                  Password Policy
                </label>
                <Select
                  defaultValue="strong"
                  onChange={() => {}}
                  options={[
                    { value: 'basic', label: 'Basic' },
                    { value: 'strong', label: 'Strong' },
                    { value: 'very-strong', label: 'Very Strong' },
                  ]}
                  className="max-w-xs"
                />
              </div>
              <Button variant="primary">Save Settings</Button>
            </div>
          </CardBody>
        </Card>
      )}

      {/* Integrations Tab */}
      {activeTab === 'integrations' && (
        <Card>
          <CardHeader>
            <CardTitle>External Integrations</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="space-y-4">
              {[
                { name: 'Email Service', status: 'connected', icon: '📧' },
                { name: 'SMS Gateway', status: 'connected', icon: '📱' },
                { name: 'Document Storage', status: 'connected', icon: '📁' },
                { name: 'Analytics Platform', status: 'disconnected', icon: '📊' },
              ].map((integration) => (
                <div key={integration.name} className="flex items-center justify-between p-4 bg-slate-50 rounded-lg">
                  <div className="flex items-center gap-4">
                    <span className="text-2xl">{integration.icon}</span>
                    <div>
                      <p className="font-medium text-slate-900">{integration.name}</p>
                      <p className={`text-sm ${integration.status === 'connected' ? 'text-green-600' : 'text-slate-500'}`}>
                        {integration.status === 'connected' ? 'Connected' : 'Not Connected'}
                      </p>
                    </div>
                  </div>
                  <Button variant={integration.status === 'connected' ? 'secondary' : 'primary'} size="sm">
                    {integration.status === 'connected' ? 'Configure' : 'Connect'}
                  </Button>
                </div>
              ))}
            </div>
          </CardBody>
        </Card>
      )}
    </div>
  );
};

export default Admin;
