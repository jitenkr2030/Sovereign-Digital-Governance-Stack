import { useState, useEffect } from 'react';
import { FileText, Download, Plus, Calendar, Filter, Clock, CheckCircle, XCircle, AlertCircle, FileJson, FileSpreadsheet } from 'lucide-react';
import { Card, CardHeader, CardTitle, Button, DataTable, Modal, StatsCard } from '../components/common';
import { Report, ReportType, ReportStatus } from '../types';

// Mock data
const mockReports: Report[] = [
  {
    id: '1',
    title: 'Monthly Compliance Report - January 2024',
    type: 'compliance',
    format: 'pdf',
    status: 'completed',
    parameters: { startDate: '2024-01-01', endDate: '2024-01-31' },
    filePath: '/reports/compliance-jan-2024.pdf',
    fileSize: 2457600,
    generatedBy: 'system',
    createdAt: '2024-02-01T10:00:00Z',
    completedAt: '2024-02-01T10:05:00Z',
  },
  {
    id: '2',
    title: 'Energy Consumption Analysis Q4 2023',
    type: 'energy',
    format: 'xlsx',
    status: 'completed',
    parameters: { startDate: '2023-10-01', endDate: '2023-12-31' },
    filePath: '/reports/energy-q4-2023.xlsx',
    fileSize: 1048576,
    generatedBy: 'admin@csic.gov',
    createdAt: '2024-01-15T08:00:00Z',
    completedAt: '2024-01-15T08:03:00Z',
  },
  {
    id: '3',
    title: 'Quarterly Financial Summary',
    type: 'financial',
    format: 'pdf',
    status: 'processing',
    parameters: { startDate: '2023-10-01', endDate: '2023-12-31' },
    generatedBy: 'regulator@csic.gov',
    createdAt: '2024-02-01T14:00:00Z',
  },
  {
    id: '4',
    title: 'License Audit Trail Export',
    type: 'audit',
    format: 'csv',
    status: 'failed',
    parameters: { entityIds: ['VAS-001', 'VAS-002'] },
    error: 'Timeout: Report generation exceeded 5 minutes',
    generatedBy: 'auditor@csic.gov',
    createdAt: '2024-01-28T16:00:00Z',
  },
  {
    id: '5',
    title: 'Regulatory Filing - Annual 2023',
    type: 'regulatory',
    format: 'pdf',
    status: 'completed',
    parameters: { year: 2023 },
    filePath: '/reports/regulatory-annual-2023.pdf',
    fileSize: 5242880,
    generatedBy: 'system',
    createdAt: '2024-01-20T00:00:00Z',
    completedAt: '2024-01-20T00:10:00Z',
  },
];

const mockTemplates = [
  { id: '1', name: 'Monthly Compliance Report', type: 'compliance', description: 'Standard monthly compliance summary' },
  { id: '2', name: 'Energy Audit Report', type: 'energy', description: 'Detailed energy consumption analysis' },
  { id: '3', name: 'Financial Summary', type: 'financial', description: 'Quarterly financial overview' },
  { id: '4', name: 'License Status Report', type: 'audit', description: 'Current license status and history' },
  { id: '5', name: 'Regulatory Filing', type: 'regulatory', description: 'Official regulatory submission' },
];

const statusIcons: Record<ReportStatus, React.ReactNode> = {
  completed: <CheckCircle className="w-4 h-4 text-green-500" />,
  processing: <Clock className="w-4 h-4 text-blue-500 animate-spin" />,
  failed: <XCircle className="w-4 h-4 text-red-500" />,
  pending: <Clock className="w-4 h-4 text-yellow-500" />,
  cancelled: <XCircle className="w-4 h-4 text-gray-500" />,
};

const formatIcons: Record<string, React.ReactNode> = {
  pdf: <FileText className="w-4 h-4" />,
  csv: <FileText className="w-4 h-4" />,
  xlsx: <FileSpreadsheet className="w-4 h-4" />,
  json: <FileJson className="w-4 h-4" />,
};

export default function Reports() {
  const [reports, setReports] = useState<Report[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [showGenerateModal, setShowGenerateModal] = useState(false);
  const [selectedReport, setSelectedReport] = useState<Report | null>(null);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [typeFilter, setTypeFilter] = useState<string[]>([]);
  const [statusFilter, setStatusFilter] = useState<string[]>([]);

  // Form state
  const [generateForm, setGenerateForm] = useState({
    title: '',
    type: 'compliance' as ReportType,
    format: 'pdf' as const,
    startDate: '',
    endDate: '',
  });

  useEffect(() => {
    const fetchReports = async () => {
      setIsLoading(true);
      setTimeout(() => {
        setReports(mockReports);
        setIsLoading(false);
      }, 800);
    };

    fetchReports();
  }, []);

  const filteredReports = reports.filter((report) => {
    const matchesType = typeFilter.length === 0 || typeFilter.includes(report.type);
    const matchesStatus = statusFilter.length === 0 || statusFilter.includes(report.status);
    return matchesType && matchesStatus;
  });

  const handleGenerateReport = () => {
    // In production, this would call the API
    console.log('Generating report:', generateForm);
    setShowGenerateModal(false);
    setGenerateForm({
      title: '',
      type: 'compliance',
      format: 'pdf',
      startDate: '',
      endDate: '',
    });
  };

  const handleDownload = (report: Report) => {
    // In production, this would download the file
    console.log('Downloading report:', report.id);
  };

  const columns = [
    {
      key: 'title',
      header: 'Report Title',
      sortable: true,
      render: (row: Report) => (
        <div className="flex items-center gap-2">
          <span className="p-1 rounded bg-dark-100 dark:bg-dark-800">
            {formatIcons[row.format]}
          </span>
          <span className="font-medium">{row.title}</span>
        </div>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      sortable: true,
      render: (row: Report) => (
        <span className="badge badge-info capitalize">{row.type}</span>
      ),
    },
    {
      key: 'format',
      header: 'Format',
      sortable: true,
      render: (row: Report) => (
        <span className="uppercase text-xs font-mono">{row.format}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row: Report) => (
        <div className="flex items-center gap-2">
          {statusIcons[row.status]}
          <span className="capitalize">{row.status}</span>
        </div>
      ),
    },
    {
      key: 'createdAt',
      header: 'Created',
      sortable: true,
      render: (row: Report) => (
        <div>
          <p className="text-sm">{new Date(row.createdAt).toLocaleDateString()}</p>
          <p className="text-xs text-dark-500">{new Date(row.createdAt).toLocaleTimeString()}</p>
        </div>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'center' as const,
      render: (row: Report) => (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => {
              setSelectedReport(row);
              setShowDetailModal(true);
            }}
            className="p-1.5 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-700 transition-colors"
            title="View Details"
          >
            <FileText className="w-4 h-4" />
          </button>
          {row.status === 'completed' && (
            <button
              onClick={() => handleDownload(row)}
              className="p-1.5 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-700 transition-colors"
              title="Download"
            >
              <Download className="w-4 h-4" />
            </button>
          )}
        </div>
      ),
    },
  ];

  const stats = {
    total: reports.length,
    completed: reports.filter((r) => r.status === 'completed').length,
    processing: reports.filter((r) => r.status === 'processing').length,
    failed: reports.filter((r) => r.status === 'failed').length,
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-50">Regulatory Reports</h1>
          <p className="text-dark-500 dark:text-dark-400 mt-1">
            Generate and manage compliance reports
          </p>
        </div>
        <Button leftIcon={<Plus className="w-4 h-4" />} onClick={() => setShowGenerateModal(true)}>
          Generate Report
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <StatsCard
          title="Total Reports"
          value={stats.total}
          icon={<FileText className="w-6 h-6 text-primary-600" />}
          iconBg="bg-primary-100 dark:bg-primary-900/30"
        />
        <StatsCard
          title="Completed"
          value={stats.completed}
          icon={<CheckCircle className="w-6 h-6 text-green-600" />}
          iconBg="bg-green-100 dark:bg-green-900/30"
        />
        <StatsCard
          title="Processing"
          value={stats.processing}
          icon={<Clock className="w-6 h-6 text-blue-600" />}
          iconBg="bg-blue-100 dark:bg-blue-900/30"
        />
        <StatsCard
          title="Failed"
          value={stats.failed}
          icon={<AlertCircle className="w-6 h-6 text-red-600" />}
          iconBg="bg-red-100 dark:bg-red-900/30"
        />
      </div>

      {/* Filters */}
      <Card padding="sm">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-dark-400" />
            <select
              value={typeFilter.join(',')}
              onChange={(e) => setTypeFilter(e.target.value ? e.target.value.split(',') : [])}
              className="px-3 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
            >
              <option value="">All Types</option>
              <option value="compliance">Compliance</option>
              <option value="energy">Energy</option>
              <option value="financial">Financial</option>
              <option value="audit">Audit</option>
              <option value="regulatory">Regulatory</option>
            </select>
            <select
              value={statusFilter.join(',')}
              onChange={(e) => setStatusFilter(e.target.value ? e.target.value.split(',') : [])}
              className="px-3 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
            >
              <option value="">All Statuses</option>
              <option value="completed">Completed</option>
              <option value="processing">Processing</option>
              <option value="pending">Pending</option>
              <option value="failed">Failed</option>
            </select>
          </div>
        </div>
      </Card>

      {/* Reports Table */}
      <Card>
        <DataTable
          data={filteredReports}
          columns={columns}
          keyExtractor={(row) => row.id}
          pageSize={10}
          loading={isLoading}
          emptyMessage="No reports found"
        />
      </Card>

      {/* Generate Report Modal */}
      <Modal
        isOpen={showGenerateModal}
        onClose={() => setShowGenerateModal(false)}
        title="Generate New Report"
        size="lg"
      >
        <div className="space-y-6">
          <div>
            <label className="label">Report Title</label>
            <input
              type="text"
              value={generateForm.title}
              onChange={(e) => setGenerateForm({ ...generateForm, title: e.target.value })}
              placeholder="Enter report title..."
              className="input"
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="label">Report Type</label>
              <select
                value={generateForm.type}
                onChange={(e) => setGenerateForm({ ...generateForm, type: e.target.value as ReportType })}
                className="input"
              >
                <option value="compliance">Compliance Report</option>
                <option value="energy">Energy Report</option>
                <option value="financial">Financial Report</option>
                <option value="audit">Audit Report</option>
                <option value="regulatory">Regulatory Report</option>
              </select>
            </div>
            <div>
              <label className="label">Output Format</label>
              <select
                value={generateForm.format}
                onChange={(e) => setGenerateForm({ ...generateForm, format: e.target.value as typeof generateForm.format })}
                className="input"
              >
                <option value="pdf">PDF Document</option>
                <option value="xlsx">Excel Spreadsheet</option>
                <option value="csv">CSV File</option>
                <option value="json">JSON Data</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="label">Start Date</label>
              <input
                type="date"
                value={generateForm.startDate}
                onChange={(e) => setGenerateForm({ ...generateForm, startDate: e.target.value })}
                className="input"
              />
            </div>
            <div>
              <label className="label">End Date</label>
              <input
                type="date"
                value={generateForm.endDate}
                onChange={(e) => setGenerateForm({ ...generateForm, endDate: e.target.value })}
                className="input"
              />
            </div>
          </div>

          <div>
            <label className="label">Quick Templates</label>
            <div className="grid grid-cols-2 gap-2 mt-2">
              {mockTemplates.map((template) => (
                <button
                  key={template.id}
                  onClick={() =>
                    setGenerateForm({
                      ...generateForm,
                      title: template.name,
                      type: template.type as ReportType,
                    })
                  }
                  className={`p-3 text-left rounded-lg border transition-colors ${
                    generateForm.type === template.type
                      ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                      : 'border-dark-200 dark:border-dark-700 hover:border-primary-300'
                  }`}
                >
                  <p className="font-medium text-sm">{template.name}</p>
                  <p className="text-xs text-dark-500">{template.description}</p>
                </button>
              ))}
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t">
            <Button variant="secondary" onClick={() => setShowGenerateModal(false)}>
              Cancel
            </Button>
            <Button onClick={handleGenerateReport} disabled={!generateForm.title}>
              Generate Report
            </Button>
          </div>
        </div>
      </Modal>

      {/* Report Detail Modal */}
      <Modal
        isOpen={showDetailModal}
        onClose={() => setShowDetailModal(false)}
        title="Report Details"
        size="md"
      >
        {selectedReport && (
          <div className="space-y-6">
            <div className="flex items-start justify-between">
              <div className="flex items-center gap-3">
                {formatIcons[selectedReport.format]}
                <div>
                  <h3 className="font-semibold">{selectedReport.title}</h3>
                  <p className="text-sm text-dark-500">ID: {selectedReport.id}</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {statusIcons[selectedReport.status]}
                <span className="capitalize">{selectedReport.status}</span>
              </div>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-dark-500">Type</p>
                <p className="font-medium capitalize">{selectedReport.type}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Format</p>
                <p className="font-medium uppercase">{selectedReport.format}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Created</p>
                <p className="font-medium">{new Date(selectedReport.createdAt).toLocaleString()}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Completed</p>
                <p className="font-medium">
                  {selectedReport.completedAt
                    ? new Date(selectedReport.completedAt).toLocaleString()
                    : '-'}
                </p>
              </div>
              {selectedReport.fileSize && (
                <div>
                  <p className="text-sm text-dark-500">File Size</p>
                  <p className="font-medium">
                    {(selectedReport.fileSize / 1024 / 1024).toFixed(2)} MB
                  </p>
                </div>
              )}
              {selectedReport.error && (
                <div className="col-span-2">
                  <p className="text-sm text-dark-500">Error</p>
                  <p className="font-medium text-red-600">{selectedReport.error}</p>
                </div>
              )}
            </div>

            {selectedReport.status === 'completed' && (
              <div className="flex justify-end gap-3 pt-4 border-t">
                <Button variant="secondary" onClick={() => setShowDetailModal(false)}>
                  Close
                </Button>
                <Button leftIcon={<Download className="w-4 h-4" />} onClick={() => handleDownload(selectedReport)}>
                  Download Report
                </Button>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}
