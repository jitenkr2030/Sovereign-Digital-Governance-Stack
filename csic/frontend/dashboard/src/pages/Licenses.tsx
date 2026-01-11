import { useState, useEffect } from 'react';
import { Search, Filter, Download, Plus, Eye, Edit, Trash2, CheckCircle, XCircle, AlertTriangle } from 'lucide-react';
import { Card, CardHeader, CardTitle, Button, DataTable, Modal } from '../components/common';
import { License, LicenseStatus } from '../types';

// Mock data
const mockLicenses: License[] = [
  {
    id: '1',
    entityId: 'VAS-001',
    entityName: 'CryptoCorp Exchange',
    entityType: 'VASP',
    licenseType: 'exchange',
    licenseNumber: 'CSIC-2024-001',
    status: 'active',
    issueDate: '2024-01-15',
    expiryDate: '2025-01-15',
    jurisdiction: 'United States',
    regulatoryBody: 'FinCEN',
    conditions: ['AML/KYC Compliance', 'Transaction Reporting'],
    restrictions: [],
    complianceScore: 95,
    lastAuditDate: '2024-06-01',
    nextAuditDate: '2024-12-01',
    documents: [],
    history: [],
    createdAt: '2024-01-15T00:00:00Z',
    updatedAt: '2024-06-01T00:00:00Z',
  },
  {
    id: '2',
    entityId: 'VAS-002',
    entityName: 'SecureVault CASP',
    entityType: 'CASP',
    licenseType: 'custodian',
    licenseNumber: 'CSIC-2024-002',
    status: 'active',
    issueDate: '2024-02-01',
    expiryDate: '2025-02-01',
    jurisdiction: 'European Union',
    regulatoryBody: 'AMF',
    conditions: ['Segregation of Funds', 'Insurance Coverage'],
    restrictions: ['Limited to EU jurisdictions'],
    complianceScore: 92,
    lastAuditDate: '2024-05-15',
    nextAuditDate: '2024-11-15',
    documents: [],
    history: [],
    createdAt: '2024-02-01T00:00:00Z',
    updatedAt: '2024-05-15T00:00:00Z',
  },
  {
    id: '3',
    entityId: 'VAS-003',
    entityName: 'MiningCo Operations',
    entityType: 'VASP',
    licenseType: 'mining',
    licenseNumber: 'CSIC-2024-003',
    status: 'pending',
    issueDate: '',
    expiryDate: '',
    jurisdiction: 'Canada',
    regulatoryBody: 'FINTRAC',
    conditions: [],
    restrictions: [],
    complianceScore: 78,
    documents: [],
    history: [],
    createdAt: '2024-06-01T00:00:00Z',
    updatedAt: '2024-06-01T00:00:00Z',
  },
  {
    id: '4',
    entityId: 'VAS-004',
    entityName: 'WalletTech Ltd',
    entityType: 'VASP',
    licenseType: 'wallet',
    licenseNumber: 'CSIC-2024-004',
    status: 'suspended',
    issueDate: '2024-01-01',
    expiryDate: '2024-12-31',
    jurisdiction: 'United Kingdom',
    regulatoryBody: 'FCA',
    conditions: [],
    restrictions: ['Suspended pending investigation'],
    complianceScore: 65,
    lastAuditDate: '2024-03-01',
    nextAuditDate: '2024-09-01',
    documents: [],
    history: [],
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-07-01T00:00:00Z',
  },
  {
    id: '5',
    entityId: 'VAS-005',
    entityName: 'DeFi Protocols Inc',
    entityType: 'OTHER',
    licenseType: 'other',
    licenseNumber: 'CSIC-2024-005',
    status: 'active',
    issueDate: '2024-03-01',
    expiryDate: '2025-03-01',
    jurisdiction: 'Singapore',
    regulatoryBody: 'MAS',
    conditions: ['Quarterly Reporting', 'Smart Contract Audit'],
    restrictions: [],
    complianceScore: 88,
    lastAuditDate: '2024-06-15',
    nextAuditDate: '2024-12-15',
    documents: [],
    history: [],
    createdAt: '2024-03-01T00:00:00Z',
    updatedAt: '2024-06-15T00:00:00Z',
  },
];

const statusColors: Record<LicenseStatus, { bg: string; text: string }> = {
  active: { bg: 'bg-green-100 dark:bg-green-900/30', text: 'text-green-800 dark:text-green-400' },
  pending: { bg: 'bg-yellow-100 dark:bg-yellow-900/30', text: 'text-yellow-800 dark:text-yellow-400' },
  suspended: { bg: 'bg-red-100 dark:bg-red-900/30', text: 'text-red-800 dark:text-red-400' },
  revoked: { bg: 'bg-gray-100 dark:bg-gray-900/30', text: 'text-gray-800 dark:text-gray-400' },
  expired: { bg: 'bg-orange-100 dark:bg-orange-900/30', text: 'text-orange-800 dark:text-orange-400' },
  under_review: { bg: 'bg-blue-100 dark:bg-blue-900/30', text: 'text-blue-800 dark:text-blue-400' },
};

export default function Licenses() {
  const [licenses, setLicenses] = useState<License[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedLicense, setSelectedLicense] = useState<License | null>(null);
  const [showDetailModal, setShowDetailModal] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string[]>([]);

  useEffect(() => {
    const fetchLicenses = async () => {
      setIsLoading(true);
      // Simulate API call
      setTimeout(() => {
        setLicenses(mockLicenses);
        setIsLoading(false);
      }, 800);
    };

    fetchLicenses();
  }, []);

  const filteredLicenses = licenses.filter((license) => {
    const matchesSearch =
      license.entityName.toLowerCase().includes(searchQuery.toLowerCase()) ||
      license.licenseNumber.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesStatus = statusFilter.length === 0 || statusFilter.includes(license.status);
    return matchesSearch && matchesStatus;
  });

  const handleViewLicense = (license: License) => {
    setSelectedLicense(license);
    setShowDetailModal(true);
  };

  const columns = [
    {
      key: 'entityName',
      header: 'Entity Name',
      sortable: true,
      render: (row: License) => (
        <div>
          <p className="font-medium">{row.entityName}</p>
          <p className="text-xs text-dark-500">{row.entityId}</p>
        </div>
      ),
    },
    {
      key: 'entityType',
      header: 'Type',
      sortable: true,
      render: (row: License) => (
        <span className="badge badge-neutral">{row.entityType}</span>
      ),
    },
    {
      key: 'licenseType',
      header: 'License Type',
      sortable: true,
      render: (row: License) => (
        <span className="capitalize">{row.licenseType}</span>
      ),
    },
    {
      key: 'status',
      header: 'Status',
      sortable: true,
      render: (row: License) => (
        <span className={`badge ${statusColors[row.status]?.bg} ${statusColors[row.status]?.text}`}>
          {row.status.replace('_', ' ')}
        </span>
      ),
    },
    {
      key: 'complianceScore',
      header: 'Compliance',
      sortable: true,
      render: (row: License) => (
        <div className="flex items-center gap-2">
          <div className="w-16 h-2 bg-dark-200 dark:bg-dark-700 rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full ${
                row.complianceScore >= 90
                  ? 'bg-green-500'
                  : row.complianceScore >= 70
                  ? 'bg-yellow-500'
                  : 'bg-red-500'
              }`}
              style={{ width: `${row.complianceScore}%` }}
            />
          </div>
          <span className="text-sm font-medium">{row.complianceScore}%</span>
        </div>
      ),
    },
    {
      key: 'jurisdiction',
      header: 'Jurisdiction',
      sortable: true,
    },
    {
      key: 'expiryDate',
      header: 'Expiry',
      sortable: true,
      render: (row: License) => row.expiryDate || '-',
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'center' as const,
      render: (row: License) => (
        <div className="flex items-center justify-center gap-2">
          <button
            onClick={() => handleViewLicense(row)}
            className="p-1.5 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-700 transition-colors"
            title="View Details"
          >
            <Eye className="w-4 h-4" />
          </button>
          <button
            className="p-1.5 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-700 transition-colors"
            title="Edit"
          >
            <Edit className="w-4 h-4" />
          </button>
        </div>
      ),
    },
  ];

  const stats = {
    total: licenses.length,
    active: licenses.filter((l) => l.status === 'active').length,
    pending: licenses.filter((l) => l.status === 'pending').length,
    suspended: licenses.filter((l) => l.status === 'suspended').length,
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-dark-900 dark:text-dark-50">License Management</h1>
          <p className="text-dark-500 dark:text-dark-400 mt-1">
            Manage VASP/CASP licenses and monitor compliance status
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="secondary" leftIcon={<Download className="w-4 h-4" />}>
            Export
          </Button>
          <Button leftIcon={<Plus className="w-4 h-4" />}>
            New License
          </Button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-primary-100 dark:bg-primary-900/30">
              <Shield className="w-5 h-5 text-primary-600" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.total}</p>
              <p className="text-sm text-dark-500">Total Licenses</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-green-100 dark:bg-green-900/30">
              <CheckCircle className="w-5 h-5 text-green-600" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.active}</p>
              <p className="text-sm text-dark-500">Active</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-yellow-100 dark:bg-yellow-900/30">
              <AlertTriangle className="w-5 h-5 text-yellow-600" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.pending}</p>
              <p className="text-sm text-dark-500">Pending</p>
            </div>
          </div>
        </Card>
        <Card padding="sm">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-lg bg-red-100 dark:bg-red-900/30">
              <XCircle className="w-5 h-5 text-red-600" />
            </div>
            <div>
              <p className="text-2xl font-bold">{stats.suspended}</p>
              <p className="text-sm text-dark-500">Suspended</p>
            </div>
          </div>
        </Card>
      </div>

      {/* Filters */}
      <Card padding="sm">
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-dark-400" />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search by entity name or license number..."
              className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
            />
          </div>
          <div className="flex items-center gap-2">
            <Filter className="w-4 h-4 text-dark-400" />
            <select
              value={statusFilter.join(',')}
              onChange={(e) => setStatusFilter(e.target.value ? e.target.value.split(',') : [])}
              className="px-3 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-sm"
            >
              <option value="">All Statuses</option>
              <option value="active">Active</option>
              <option value="pending">Pending</option>
              <option value="suspended">Suspended</option>
              <option value="revoked">Revoked</option>
            </select>
          </div>
        </div>
      </Card>

      {/* Data Table */}
      <Card>
        <DataTable
          data={filteredLicenses}
          columns={columns}
          keyExtractor={(row) => row.id}
          pageSize={10}
          loading={isLoading}
          emptyMessage="No licenses found"
          onRowClick={handleViewLicense}
        />
      </Card>

      {/* License Detail Modal */}
      <Modal
        isOpen={showDetailModal}
        onClose={() => setShowDetailModal(false)}
        title="License Details"
        size="lg"
      >
        {selectedLicense && (
          <div className="space-y-6">
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-lg font-semibold">{selectedLicense.entityName}</h3>
                <p className="text-sm text-dark-500">{selectedLicense.licenseNumber}</p>
              </div>
              <span className={`badge ${statusColors[selectedLicense.status]?.bg} ${statusColors[selectedLicense.status]?.text}`}>
                {selectedLicense.status.replace('_', ' ')}
              </span>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-dark-500">Entity Type</p>
                <p className="font-medium">{selectedLicense.entityType}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">License Type</p>
                <p className="font-medium capitalize">{selectedLicense.licenseType}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Jurisdiction</p>
                <p className="font-medium">{selectedLicense.jurisdiction}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Regulatory Body</p>
                <p className="font-medium">{selectedLicense.regulatoryBody}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Issue Date</p>
                <p className="font-medium">{selectedLicense.issueDate || '-'}</p>
              </div>
              <div>
                <p className="text-sm text-dark-500">Expiry Date</p>
                <p className="font-medium">{selectedLicense.expiryDate || '-'}</p>
              </div>
            </div>

            <div>
              <p className="text-sm text-dark-500 mb-2">Compliance Score</p>
              <div className="flex items-center gap-3">
                <div className="flex-1 h-3 bg-dark-200 dark:bg-dark-700 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full ${
                      selectedLicense.complianceScore >= 90
                        ? 'bg-green-500'
                        : selectedLicense.complianceScore >= 70
                        ? 'bg-yellow-500'
                        : 'bg-red-500'
                    }`}
                    style={{ width: `${selectedLicense.complianceScore}%` }}
                  />
                </div>
                <span className="font-semibold">{selectedLicense.complianceScore}%</span>
              </div>
            </div>

            {selectedLicense.conditions.length > 0 && (
              <div>
                <p className="text-sm text-dark-500 mb-2">Conditions</p>
                <ul className="list-disc list-inside space-y-1">
                  {selectedLicense.conditions.map((condition, index) => (
                    <li key={index} className="text-sm">{condition}</li>
                  ))}
                </ul>
              </div>
            )}

            {selectedLicense.restrictions.length > 0 && (
              <div>
                <p className="text-sm text-dark-500 mb-2">Restrictions</p>
                <ul className="list-disc list-inside space-y-1">
                  {selectedLicense.restrictions.map((restriction, index) => (
                    <li key={index} className="text-sm text-yellow-600">{restriction}</li>
                  ))}
                </ul>
              </div>
            )}

            <div className="flex justify-end gap-3 pt-4 border-t">
              <Button variant="secondary" onClick={() => setShowDetailModal(false)}>
                Close
              </Button>
              <Button>Edit License</Button>
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
