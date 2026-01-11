import React, { useState } from 'react';
import { Card, Button, Select, Input } from '../common';
import { ComplianceReport } from '../../services/compliance';

interface ComplianceReportGeneratorProps {
  onGenerate: (type: ComplianceReport['type'], period: { start: string; end: string }, categoryId?: string) => Promise<void>;
  onDownload: (reportId: string) => Promise<void>;
  onView: (report: ComplianceReport) => void;
  recentReports: ComplianceReport[];
  loading?: boolean;
}

export const ComplianceReportGenerator: React.FC<ComplianceReportGeneratorProps> = ({
  onGenerate,
  onDownload,
  onView,
  recentReports,
  loading = false,
}) => {
  const [reportType, setReportType] = useState<ComplianceReport['type']>('summary');
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [showHistory, setShowHistory] = useState(false);

  const reportTypes = [
    { value: 'summary', label: 'Summary Report', description: 'Overview of compliance status' },
    { value: 'full', label: 'Full Report', description: 'Detailed compliance analysis' },
    { value: 'category', label: 'Category Report', description: 'Report for specific category' },
    { value: 'violation', label: 'Violation Report', description: 'Report of all violations' },
  ];

  const handleGenerate = async (e: React.FormEvent) => {
    e.preventDefault();
    await onGenerate(reportType, { start: startDate, end: endDate }, categoryId || undefined);
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getTypeLabel = (type: ComplianceReport['type']) => {
    return reportTypes.find((t) => t.value === type)?.label || type;
  };

  return (
    <div className="space-y-6">
      {/* Report Generator Form */}
      <Card>
        <CardBody>
          <h2 className="text-lg font-semibold text-slate-900 mb-4">Generate Report</h2>
          <form onSubmit={handleGenerate} className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Select
                label="Report Type"
                value={reportType}
                onChange={(e) => setReportType(e.target.value as typeof reportType)}
                options={reportTypes.map((t) => ({ value: t.value, label: t.label }))}
                fullWidth
              />
              <Select
                label="Category (Optional)"
                value={categoryId}
                onChange={(e) => setCategoryId(e.target.value)}
                options={[
                  { value: '', label: 'All Categories' },
                  { value: 'safety', label: 'Safety' },
                  { value: 'environmental', label: 'Environmental' },
                  { value: 'operational', label: 'Operational' },
                  { value: 'financial', label: 'Financial' },
                ]}
                fullWidth
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Input
                type="date"
                label="Start Date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                required
                fullWidth
              />
              <Input
                type="date"
                label="End Date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                required
                fullWidth
              />
            </div>

            <div className="flex justify-end pt-4">
              <Button
                type="submit"
                variant="primary"
                loading={loading}
                disabled={!startDate || !endDate}
              >
                Generate Report
              </Button>
            </div>
          </form>
        </CardBody>
      </Card>

      {/* Recent Reports */}
      <Card>
        <CardBody>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold text-slate-900">Recent Reports</h2>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setShowHistory(!showHistory)}
            >
              {showHistory ? 'Hide History' : 'View All'}
            </Button>
          </div>

          {recentReports.length > 0 ? (
            <div className="space-y-3">
              {(showHistory ? recentReports : recentReports.slice(0, 5)).map((report) => (
                <div
                  key={report.id}
                  className="flex items-center justify-between p-4 bg-slate-50 rounded-lg hover:bg-slate-100 transition-colors"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-blue-100 rounded-lg">
                      <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    </div>
                    <div>
                      <p className="font-medium text-slate-900">{report.title}</p>
                      <div className="flex items-center gap-2 text-sm text-slate-500">
                        <span>{getTypeLabel(report.type)}</span>
                        <span>•</span>
                        <span>{formatDate(report.generatedAt)}</span>
                        <span>•</span>
                        <span>By {report.generatedBy}</span>
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <Button
                      variant="secondary"
                      size="sm"
                      onClick={() => onView(report)}
                    >
                      View
                    </Button>
                    <Button
                      variant="primary"
                      size="sm"
                      onClick={() => onDownload(report.id)}
                    >
                      Download
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <svg className="w-12 h-12 text-slate-300 mx-auto mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p className="text-slate-500">No reports generated yet</p>
              <p className="text-sm text-slate-400 mt-1">Generate your first report using the form above</p>
            </div>
          )}
        </CardBody>
      </Card>
    </div>
  );
};

interface AuditTrailViewerProps {
  entries: {
    id: string;
    action: string;
    entityType: string;
    entityName: string;
    performedBy: string;
    performedAt: string;
    details: string;
    previousValue?: string;
    newValue?: string;
  }[];
  onExport: (format: 'csv' | 'pdf' | 'json') => void;
  onFilter: (filter: Record<string, string>) => void;
  loading?: boolean;
}

export const AuditTrailViewer: React.FC<AuditTrailViewerProps> = ({
  entries,
  onExport,
  onFilter,
  loading = false,
}) => {
  const [actionFilter, setActionFilter] = useState('all');
  const [entityFilter, setEntityFilter] = useState('all');

  const actionTypes = [...new Set(entries.map((e) => e.action))];
  const entityTypes = [...new Set(entries.map((e) => e.entityType))];

  const handleFilterChange = () => {
    const filter: Record<string, string> = {};
    if (actionFilter !== 'all') filter.action = actionFilter;
    if (entityFilter !== 'all') filter.entityType = entityFilter;
    onFilter(filter);
  };

  React.useEffect(() => {
    handleFilterChange();
  }, [actionFilter, entityFilter]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getActionColor = (action: string) => {
    if (action.includes('create') || action.includes('add')) return 'text-green-600 bg-green-100';
    if (action.includes('update') || action.includes('modify')) return 'text-blue-600 bg-blue-100';
    if (action.includes('delete') || action.includes('remove')) return 'text-red-600 bg-red-100';
    return 'text-slate-600 bg-slate-100';
  };

  return (
    <Card>
      <CardBody>
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-slate-900">Audit Trail</h2>
          <div className="flex items-center gap-2">
            <select
              value={actionFilter}
              onChange={(e) => setActionFilter(e.target.value)}
              className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="all">All Actions</option>
              {actionTypes.map((action) => (
                <option key={action} value={action}>
                  {action}
                </option>
              ))}
            </select>
            <select
              value={entityFilter}
              onChange={(e) => setEntityFilter(e.target.value)}
              className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="all">All Entities</option>
              {entityTypes.map((entity) => (
                <option key={entity} value={entity}>
                  {entity}
                </option>
              ))}
            </select>
            <div className="flex items-center gap-1">
              <Button variant="secondary" size="sm" onClick={() => onExport('csv')}>
                CSV
              </Button>
              <Button variant="secondary" size="sm" onClick={() => onExport('json')}>
                JSON
              </Button>
              <Button variant="secondary" size="sm" onClick={() => onExport('pdf')}>
                PDF
              </Button>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          {entries.map((entry) => (
            <div
              key={entry.id}
              className="p-4 bg-slate-50 rounded-lg hover:bg-slate-100 transition-colors"
            >
              <div className="flex items-start justify-between">
                <div className="flex items-start gap-3">
                  <div className="flex-1">
                    <div className="flex items-center gap-2 mb-1">
                      <span className={`px-2 py-0.5 text-xs font-medium rounded-full ${getActionColor(entry.action)}`}>
                        {entry.action}
                      </span>
                      <span className="text-sm text-slate-500">{entry.entityType}</span>
                    </div>
                    <p className="font-medium text-slate-900">{entry.entityName}</p>
                    <p className="text-sm text-slate-500 mt-1">{entry.details}</p>
                    {(entry.previousValue || entry.newValue) && (
                      <div className="flex items-center gap-2 mt-2 text-sm">
                        <span className="text-slate-400 line-through">{entry.previousValue}</span>
                        <svg className="w-4 h-4 text-slate-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8l4 4m0 0l-4 4m4-4H3" />
                        </svg>
                        <span className="text-slate-700">{entry.newValue}</span>
                      </div>
                    )}
                  </div>
                </div>
                <div className="text-right text-sm text-slate-500">
                  <p>{entry.performedBy}</p>
                  <p>{formatDate(entry.performedAt)}</p>
                </div>
              </div>
            </div>
          ))}

          {entries.length === 0 && (
            <div className="text-center py-8">
              <p className="text-slate-500">No audit trail entries found</p>
            </div>
          )}
        </div>
      </CardBody>
    </Card>
  );
};

export default ComplianceReportGenerator;
