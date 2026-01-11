import React, { useState } from 'react';
import { Card, Button, Checkbox, Modal, Input, Textarea } from '../common';
import { ComplianceChecklist } from '../../services/compliance';

interface ChecklistViewerProps {
  checklist: ComplianceChecklist[];
  onUpdateItem: (itemId: string, status: ComplianceChecklist['status'], notes?: string) => Promise<void>;
  onViewEvidence?: (item: ComplianceChecklist) => void;
  loading?: boolean;
}

export const ChecklistViewer: React.FC<ChecklistViewerProps> = ({
  checklist,
  onUpdateItem,
  onViewEvidence,
  loading = false,
}) => {
  const [expandedItem, setExpandedItem] = useState<string | null>(null);
  const [selectedItems, setSelectedItems] = useState<(string | number)[]>([]);
  const [filter, setFilter] = useState<ComplianceChecklist['status'] | 'all'>('all');
  const [priorityFilter, setPriorityFilter] = useState<'all' | 'high' | 'medium' | 'low'>('all');
  const [showNotesModal, setShowNotesModal] = useState(false);
  const [currentItem, setCurrentItem] = useState<ComplianceChecklist | null>(null);
  const [notes, setNotes] = useState('');
  const [bulkUpdateStatus, setBulkUpdateStatus] = useState<ComplianceChecklist['status'] | null>(null);

  const filteredChecklist = checklist.filter((item) => {
    if (filter !== 'all' && item.status !== filter) return false;
    if (priorityFilter !== 'all' && item.priority !== priorityFilter) return false;
    return true;
  });

  const handleUpdateStatus = async (item: ComplianceChecklist, status: ComplianceChecklist['status']) => {
    if (status === 'pending' || status === 'not-applicable') {
      // Show notes modal for these statuses
      setCurrentItem(item);
      setShowNotesModal(true);
    } else {
      await onUpdateItem(item.id, status);
    }
  };

  const handleSaveNotes = async () => {
    if (currentItem) {
      await onUpdateItem(currentItem.id, currentItem.status === 'pending' ? 'pending' : 'not-applicable', notes);
      setShowNotesModal(false);
      setCurrentItem(null);
      setNotes('');
    }
  };

  const handleBulkUpdate = async () => {
    if (bulkUpdateStatus) {
      for (const itemId of selectedItems) {
        await onUpdateItem(itemId as string, bulkUpdateStatus);
      }
      setSelectedItems([]);
      setBulkUpdateStatus(null);
    }
  };

  const statusColors = {
    compliant: 'bg-green-100 text-green-800',
    partial: 'bg-amber-100 text-amber-800',
    'non-compliant': 'bg-red-100 text-red-800',
    pending: 'bg-slate-100 text-slate-800',
    'not-applicable': 'bg-gray-100 text-gray-800',
  };

  const priorityColors = {
    high: 'bg-red-100 text-red-700 border-red-200',
    medium: 'bg-amber-100 text-amber-700 border-amber-200',
    low: 'bg-green-100 text-green-700 border-green-200',
  };

  return (
    <div className="space-y-4">
      {/* Filters */}
      <div className="flex flex-wrap items-center gap-4">
        <div className="flex items-center space-x-2">
          <label className="text-sm text-slate-600">Status:</label>
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value as typeof filter)}
            className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All</option>
            <option value="compliant">Compliant</option>
            <option value="partial">Partial</option>
            <option value="non-compliant">Non-Compliant</option>
            <option value="pending">Pending</option>
            <option value="not-applicable">Not Applicable</option>
          </select>
        </div>
        <div className="flex items-center space-x-2">
          <label className="text-sm text-slate-600">Priority:</label>
          <select
            value={priorityFilter}
            onChange={(e) => setPriorityFilter(e.target.value as typeof priorityFilter)}
            className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="all">All</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
        <div className="flex items-center space-x-2 text-sm text-slate-500">
          <span>{filteredChecklist.length} items</span>
        </div>
      </div>

      {/* Bulk Actions */}
      {selectedItems.length > 0 && (
        <div className="flex items-center gap-4 p-3 bg-blue-50 rounded-lg border border-blue-200">
          <span className="text-sm font-medium text-blue-700">
            {selectedItems.length} selected
          </span>
          <select
            value={bulkUpdateStatus || ''}
            onChange={(e) => setBulkUpdateStatus(e.target.value as ComplianceChecklist['status'])}
            className="px-3 py-1.5 text-sm border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Set status...</option>
            <option value="compliant">Mark Compliant</option>
            <option value="partial">Mark Partial</option>
            <option value="non-compliant">Mark Non-Compliant</option>
            <option value="pending">Mark Pending</option>
          </select>
          {bulkUpdateStatus && (
            <Button variant="primary" size="sm" onClick={handleBulkUpdate}>
              Apply
            </Button>
          )}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setSelectedItems([])}
          >
            Clear
          </Button>
        </div>
      )}

      {/* Checklist Items */}
      <div className="space-y-3">
        {filteredChecklist.map((item) => (
          <Card
            key={item.id}
            className={`transition-all ${
              expandedItem === item.id ? 'ring-2 ring-blue-500' : ''
            }`}
          >
            <CardBody className="p-4">
              <div className="flex items-start gap-4">
                <Checkbox
                  checked={selectedItems.includes(item.id)}
                  onChange={(e) => {
                    if (e.target.checked) {
                      setSelectedItems([...selectedItems, item.id]);
                    } else {
                      setSelectedItems(selectedItems.filter((id) => id !== item.id));
                    }
                  }}
                />
                <div className="flex-1">
                  <div className="flex items-start justify-between">
                    <div>
                      <h3 className="font-medium text-slate-900">{item.title}</h3>
                      <p className="text-sm text-slate-500 mt-1">{item.description}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full border ${priorityColors[item.priority]}`}
                      >
                        {item.priority}
                      </span>
                      <span
                        className={`px-2 py-1 text-xs font-medium rounded-full ${statusColors[item.status]}`}
                      >
                        {item.status.replace('-', ' ')}
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center gap-4 mt-3 text-sm text-slate-500">
                    <span>Last checked: {new Date(item.lastChecked).toLocaleDateString()}</span>
                    {item.checkedBy && <span>By: {item.checkedBy}</span>}
                    {item.evidence && item.evidence.length > 0 && (
                      <button
                        onClick={() => onViewEvidence?.(item)}
                        className="text-blue-600 hover:text-blue-700"
                      >
                        {item.evidence.length} evidence(s)
                      </button>
                    )}
                  </div>

                  {item.notes && (
                    <p className="mt-2 text-sm text-slate-600 bg-slate-50 p-2 rounded">
                      {item.notes}
                    </p>
                  )}

                  {expandedItem === item.id && (
                    <div className="mt-4 pt-4 border-t border-slate-200">
                      <p className="text-sm font-medium text-slate-700 mb-2">
                        Update Status:
                      </p>
                      <div className="flex flex-wrap gap-2">
                        {(['compliant', 'partial', 'non-compliant', 'pending', 'not-applicable'] as const).map(
                          (status) => (
                            <button
                              key={status}
                              onClick={() => handleUpdateStatus(item, status)}
                              disabled={loading || item.status === status}
                              className={`px-3 py-1.5 text-sm rounded-lg border transition-colors ${
                                item.status === status
                                  ? statusColors[status]
                                  : 'bg-white border-slate-300 text-slate-700 hover:bg-slate-50'
                              } disabled:opacity-50`}
                            >
                              {status.replace('-', ' ')}
                            </button>
                          )
                        )}
                      </div>
                    </div>
                  )}
                </div>
                <button
                  onClick={() => setExpandedItem(expandedItem === item.id ? null : item.id)}
                  className="p-1 text-slate-400 hover:text-slate-600"
                >
                  <svg
                    className={`w-5 h-5 transition-transform ${
                      expandedItem === item.id ? 'rotate-180' : ''
                    }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                  </svg>
                </button>
              </div>
            </CardBody>
          </Card>
        ))}

        {filteredChecklist.length === 0 && (
          <div className="text-center py-8">
            <p className="text-slate-500">No checklist items match your filters</p>
          </div>
        )}
      </div>

      {/* Notes Modal */}
      <Modal
        isOpen={showNotesModal}
        onClose={() => setShowNotesModal(false)}
        title="Add Notes"
        size="md"
      >
        <div className="space-y-4">
          <Textarea
            label="Notes"
            value={notes}
            onChange={(e) => setNotes(e.target.value)}
            placeholder="Add notes about this status update..."
            rows={4}
          />
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowNotesModal(false)}>
              Cancel
            </Button>
            <Button variant="primary" onClick={handleSaveNotes}>
              Save
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default ChecklistViewer;
