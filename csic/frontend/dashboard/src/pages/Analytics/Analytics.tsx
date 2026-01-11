import React, { useState, useEffect } from 'react';
import { Card, CardBody, CardHeader, CardTitle, Select, Button } from '../../components/common';

const Analytics: React.FC = () => {
  const [timeRange, setTimeRange] = useState('week');

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Analytics</h1>
          <p className="text-slate-500 mt-1">Comprehensive analytics and insights</p>
        </div>
        <Select
          value={timeRange}
          onChange={(e) => setTimeRange(e.target.value)}
          options={[
            { value: 'week', label: 'Last 7 Days' },
            { value: 'month', label: 'Last 30 Days' },
            { value: 'quarter', label: 'Last Quarter' },
            { value: 'year', label: 'Last Year' },
          ]}
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Energy Consumption Trends</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="h-64 flex items-center justify-center bg-slate-50 rounded-lg">
              <p className="text-slate-500">Chart Placeholder</p>
            </div>
          </CardBody>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Compliance Scores</CardTitle>
          </CardHeader>
          <CardBody>
            <div className="h-64 flex items-center justify-center bg-slate-50 rounded-lg">
              <p className="text-slate-500">Chart Placeholder</p>
            </div>
          </CardBody>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Key Performance Indicators</CardTitle>
        </CardHeader>
        <CardBody>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center">
              <p className="text-4xl font-bold text-slate-900">87%</p>
              <p className="text-slate-500 mt-1">Average Compliance</p>
            </div>
            <div className="text-center">
              <p className="text-4xl font-bold text-slate-900">1,234</p>
              <p className="text-slate-500 mt-1">Total Inspections</p>
            </div>
            <div className="text-center">
              <p className="text-4xl font-bold text-slate-900">98.5%</p>
              <p className="text-slate-500 mt-1">Uptime Rate</p>
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};

export default Analytics;
