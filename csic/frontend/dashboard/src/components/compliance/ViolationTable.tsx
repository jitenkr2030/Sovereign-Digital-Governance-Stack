import React, { useState } from 'react';
import { Card, Button, Pagination, Modal, Input, Select } from '../common';
import { Violation } from '../../services/compliance';

interface ViolationTableProps {
  violations: Violation[];
  total: number;
  currentPage: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onUpdateViolation: (violationId: string, updates: Partial<Violation>) => Promise<void>;
  onViewDetails: (violation: Violation) => void;
  loading?: boolean;
}

export const ViolationTable: React.FC<ViolationTableProps> = ({
  violations,
  total,
  currentPage,
  pageSize,
  onPageChange,
  onUpdateViolation,
  onViewDetails,
  loading = false,
}) => {
  const [statusFilter, setStatusFilter] = useState<Violation['status'] | 'all'>('all');
  const [severityFilter, setSeverityFilter] = useState<Violation['severity'] | 'all'>('all');
  const [showUpdateModal, setShowUpdateModal] = useState(false);
  const [currentViolation, setCurrentViolation] = useState<Violation | null>(null);
  const [updateStatus, setUpdateStatus] = useState<Violation['status'] | null>(null);
  const [updateNotes, setUpdateNotes] = useState('');

  const filteredViolations = violations.filter((v) => {
    if (statusFilter !== 'all' && v.status !== statusFilter) return false;
    if (severityFilter !== 'all' && v.severity !== severityFilter) return false;
    return true;
  });

  const severityConfig = {
    critical: { color: 'bg-red-600 text-white', label: 'Critical' },
    high: { color: 'bg-orange-500 text-white', label: 'High' },
    medium: { color: 'bg-amber-500 text-white', label: 'Medium' },
    low: { color: 'bg-blue-500 text-white', label: 'Low' },
  };

  const statusConfig = {
    open: { color: 'bg-red-100 text-red-800', label: 'Open' },
    'in-progress': { color: 'bg-amber-100 text-amber-800', label: 'In Progress' },
    resolved: { color: 'bg-green-100 text-green-800', label: 'Resolved' },
    waived: { color: 'bg-slate-100 text-slate-800', label: 'Waived' },
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const handleOpenUpdate = (violation: Violation) => {
    setCurrentViolation(violation);
    setUpdateStatus(violation.status);
    setUpdateNotes('');
    setShowUpdateModal(true);
  };

  const handleUpdate = async () => {
    if (currentViolation && updateStatus) {
      await onUpdateViolation(currentViolation.id, {
        status: updateStatus,
        notes: updateNotes,
      });
      setShowUpdateModal(false);
      setCurrentViolation(null);
    }
  };

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center space-x-2">
          <label className="text-sm text-slate-600">Status:</label>
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as typeof statusFilter)}
            className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All</option>
            <option value="open">Open</option>
            <option value="in-progress">In Progress</option>
            <option value="resolved">Resolved</option>
            <option value="waived">Waived</option>
          </select>
        </div>
        <div className="flex items-center space-x-2">
          <label className="text-sm text-slate-600">Severity:</label>
          <select
            value={severityFilter}
            onChange={(e) => setSeverityFilter(e.target.value as typeof severityFilter)}
            className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All</option>
            <option value="critical">Critical</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div className="flex items-center space-x-2 text-sm text-slate-500">
          <span>{filteredViolations.length} of {total} violations</span>
        </div>
      </div>

      {/* Table */}
      <Card>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-slate-50 border-b border-slate-200">
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Violation
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Severity
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Detected
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Due Date
                </th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Assigned To
                </th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {filteredViolations.map((violation) => (
                <tr
                  key={violation.id}
                  className={`hover:bg-slate-50 transition-colors ${
                    loading ? 'opacity-50' : ''
                  }`}
                >
                  <td className="px-4 py-4">
                    <div>
                      <p className="font-medium text-slate-900">{violation.title}</p>
                      <p className="text-sm text-slate-500 truncate max-w-xs">
                        {violation.description}
                      </p>
                    </div>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${severityConfig[violation.severity].color}`}
                    >
                      {severityConfig[violation.severity].label}
                    </span>
                  </td>
                  <td className="px-4 py-4">
                    <span
                      className={`inline-flex px-2 py-1 text-xs font-medium rounded-full ${statusConfig[violation.status].color}`}
                    >
                      {statusConfig[violation.status].label}
                    </span>
                  </td>
                  <td className="px-4 py-4 text-sm text-slate-600">
                    {formatDate(violation.detectedAt)}
                  </td>
                  <td className="px-4 py-4 text-sm text-slate-600">
                    {formatDate(violation.dueDate)}
                  </td>
                  <td className="px-4 py-4 text-sm text-slate-600">
                    {violation.assignedTo || '-'}
                  </td>
                  <td className="px-4 py-4 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <Button
                        variant="secondary"
                        size="sm"
                        onClick={() => onViewDetails(violation)}
                      >
                        View
                      </Button>
                      <Button
                        variant="primary"
                        size="sm"
                        onClick={() => handleOpenUpdate(violation)}
                      >
                        Update
                      </Button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {filteredViolations.length === 0 && (
          <div className="text-center py-8">
            <p className="text-slate-500">No violations found</p>
          </div>
        )}

        {total > pageSize && (
          <div className="px-4 py-4 border-t border-slate-200">
            <Pagination
              currentPage={currentPage}
              totalPages={Math.ceil(total / pageSize)}
              totalItems={total}
              pageSize={pageSize}
              onPageChange={onPageChange}
            />
          </div>
        )}
      </Card>

      {/* Update Modal */}
      <Modal
        isOpen={showUpdateModal}
        onClose={() => setShowUpdateModal(false)}
        title="Update Violation Status"
        size="md"
      >
        <div className="space-y-4">
          {currentViolation && (
            <>
              <div className="bg-slate-50 rounded-lg p-3">
                <p className="font-medium text-slate-900">{currentViolation.title}</p>
                <p className="text-sm text-slate-500 mt-1">{currentViolation.description}</p>
              </div>

              <Select
                label="Status"
                value={updateStatus || ''}
                onChange={(e) => setUpdateStatus(e.target.value as Violation['status'])}
                options={[
                  { value: 'open', label: 'Open' },
                  { value: 'in-progress', label: 'In Progress' },
                  { value: 'resolved', label: 'Resolved' },
                  { value: 'waived', label: 'Waived' },
                ]}
              />

              <Textarea
                label="Notes"
                value={updateNotes}
                onChange={(e) => setUpdateNotes(e.target.value)}
                placeholder="Add notes about this update..."
                rows={4}
              />

              <div className="flex justify-end gap-3">
                <Button variant="secondary" onClick={() => setShowUpdateModal(false)}>
                  Cancel
                </Button>
                <Button variant="primary" onClick={handleUpdate}>
                  Update
                </Button>
              </div>
            </>
          )}
        </div>
      </Modal>
    </div>
  );
};

export default ViolationTable;
