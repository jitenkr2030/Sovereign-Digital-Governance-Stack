import React, { useState } from 'react';
import { Card, CardBody, CardHeader, CardTitle, Select, Input, Button } from '../../components/common';

const Audit: React.FC = () => {
  const [entityType, setEntityType] = useState('');
  const [action, setAction] = useState('');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Audit Logs</h1>
          <p className="text-slate-500 mt-1">System-wide audit trail and activity logs</p>
        </div>
        <Button variant="secondary">Export Logs</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Filter Logs</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <Select
              label="Entity Type"
              value={entityType}
              onChange={(e) => setEntityType(e.target.value)}
              options={[
                { value: '', label: 'All Types' },
                { value: 'user', label: 'Users' },
                { value: 'policy', label: 'Policies' },
                { value: 'license', label: 'Licenses' },
                { value: 'enforcement', label: 'Enforcement' },
              ]}
              fullWidth
            />
            <Select
              label="Action"
              value={action}
              onChange={(e) => setAction(e.target.value)}
              options={[
                { value: '', label: 'All Actions' },
                { value: 'create', label: 'Create' },
                { value: 'update', label: 'Update' },
                { value: 'delete', label: 'Delete' },
                { value: 'view', label: 'View' },
              ]}
              fullWidth
            />
            <Input
              label="Date Range"
              type="date"
              fullWidth
            />
          </div>
        </CardBody>
      </Card>

      <Card>
        <CardBody>
          <div className="space-y-4">
            {[1, 2, 3, 4, 5].map((item) => (
              <div key={item} className="flex items-center justify-between p-4 bg-slate-50 rounded-lg">
                <div className="flex items-center gap-4">
                  <div className={`p-2 rounded-lg ${item % 2 === 0 ? 'bg-green-100' : 'bg-blue-100'}`}>
                    <svg className={`w-5 h-5 ${item % 2 === 0 ? 'text-green-600' : 'text-blue-600'}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                    </svg>
                  </div>
                  <div>
                    <p className="font-medium text-slate-900">User Update</p>
                    <p className="text-sm text-slate-500">User profile was updated by admin</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm text-slate-600">john.doe@csic.gov</p>
                  <p className="text-xs text-slate-400">2 hours ago</p>
                </div>
              </div>
            ))}
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

export default Audit;
