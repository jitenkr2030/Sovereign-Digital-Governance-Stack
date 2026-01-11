import React, { useState } from 'react';
import { Card, Button, Input, Textarea, Select, Modal } from '../common';
import { ComplianceChecklist, ExemptionRequest, ExemptionPayload } from '../../services/compliance';

interface ExemptionFormProps {
  requirement: ComplianceChecklist;
  onSubmit: (payload: ExemptionPayload) => Promise<void>;
  onClose: () => void;
  loading?: boolean;
}

export const ExemptionForm: React.FC<ExemptionFormProps> = ({
  requirement,
  onSubmit,
  onClose,
  loading = false,
}) => {
  const [reason, setReason] = useState('');
  const [requestedDuration, setRequestedDuration] = useState('');
  const [supportingDocuments, setSupportingDocuments] = useState<string[]>([]);
  const [documentName, setDocumentName] = useState('');
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = () => {
    const newErrors: Record<string, string> = {};

    if (!reason.trim()) {
      newErrors.reason = 'Reason is required';
    } else if (reason.length < 20) {
      newErrors.reason = 'Reason must be at least 20 characters';
    }

    if (!requestedDuration) {
      newErrors.requestedDuration = 'Requested duration is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    const payload: ExemptionPayload = {
      requirementId: requirement.id,
      reason: reason.trim(),
      requestedDuration,
      supportingDocuments,
    };

    await onSubmit(payload);
  };

  const handleAddDocument = () => {
    if (documentName.trim()) {
      setSupportingDocuments([...supportingDocuments, documentName.trim()]);
      setDocumentName('');
    }
  };

  const handleRemoveDocument = (index: number) => {
    setSupportingDocuments(supportingDocuments.filter((_, i) => i !== index));
  };

  return (
    <Card className="max-w-2xl mx-auto">
      <CardBody>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-lg font-semibold text-slate-900">Request Exemption</h2>
            <p className="text-sm text-slate-500">
              Submit an exemption request for this compliance requirement
            </p>
          </div>
          <button
            onClick={onClose}
            className="p-2 text-slate-400 hover:text-slate-600 rounded-lg hover:bg-slate-100"
          >
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Requirement Info */}
        <div className="bg-slate-50 rounded-lg p-4 mb-6">
          <h3 className="font-medium text-slate-900">{requirement.title}</h3>
          <p className="text-sm text-slate-500 mt-1">{requirement.description}</p>
          <div className="flex items-center gap-2 mt-3">
            <span className="text-xs px-2 py-1 bg-slate-200 rounded-full text-slate-700">
              {requirement.category}
            </span>
            <span className={`text-xs px-2 py-1 rounded-full ${
              requirement.priority === 'high' ? 'bg-red-100 text-red-700' :
              requirement.priority === 'medium' ? 'bg-amber-100 text-amber-700' :
              'bg-green-100 text-green-700'
            }`}>
              {requirement.priority} priority
            </span>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <Textarea
            label="Reason for Exemption"
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder="Provide a detailed explanation of why this exemption is needed..."
            error={errors.reason}
            rows={4}
            fullWidth
          />

          <Select
            label="Requested Duration"
            value={requestedDuration}
            onChange={(e) => setRequestedDuration(e.target.value)}
            error={errors.requestedDuration}
            options={[
              { value: '', label: 'Select duration...' },
              { value: '1-month', label: '1 Month' },
              { value: '3-months', label: '3 Months' },
              { value: '6-months', label: '6 Months' },
              { value: '1-year', label: '1 Year' },
              { value: 'permanent', label: 'Permanent' },
            ]}
            fullWidth
          />

          <div>
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Supporting Documents (Optional)
            </label>
            <div className="flex gap-2 mb-2">
              <Input
                value={documentName}
                onChange={(e) => setDocumentName(e.target.value)}
                placeholder="Document name or URL..."
                fullWidth
              />
              <Button
                type="button"
                variant="secondary"
                onClick={handleAddDocument}
                disabled={!documentName.trim()}
              >
                Add
              </Button>
            </div>
            {supportingDocuments.length > 0 && (
              <ul className="space-y-2">
                {supportingDocuments.map((doc, index) => (
                  <li
                    key={index}
                    className="flex items-center justify-between bg-slate-50 px-3 py-2 rounded-lg"
                  >
                    <span className="text-sm text-slate-700">{doc}</span>
                    <button
                      type="button"
                      onClick={() => handleRemoveDocument(index)}
                      className="text-slate-400 hover:text-red-500"
                    >
                      <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start gap-3">
              <svg className="w-5 h-5 text-blue-600 mt-0.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <div>
                <p className="text-sm font-medium text-blue-800">Review Process</p>
                <p className="text-sm text-blue-600 mt-1">
                  Your exemption request will be reviewed by the compliance team.
                  You will be notified of the decision within 5-7 business days.
                </p>
              </div>
            </div>
          </div>

          <div className="flex justify-end gap-3 pt-4 border-t border-slate-200">
            <Button
              type="button"
              variant="secondary"
              onClick={onClose}
              disabled={loading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="primary"
              loading={loading}
            >
              Submit Request
            </Button>
          </div>
        </form>
      </CardBody>
    </Card>
  );
};

interface ExemptionListProps {
  exemptions: ExemptionRequest[];
  onApprove?: (exemptionId: string) => Promise<void>;
  onReject?: (exemptionId: string, reason: string) => Promise<void>;
  loading?: boolean;
}

export const ExemptionList: React.FC<ExemptionListProps> = ({
  exemptions,
  onApprove,
  onReject,
  loading = false,
}) => {
  const [showRejectModal, setShowRejectModal] = useState(false);
  const [currentExemption, setCurrentExemption] = useState<ExemptionRequest | null>(null);
  const [rejectReason, setRejectReason] = useState('');

  const statusConfig = {
    pending: { color: 'bg-amber-100 text-amber-800', label: 'Pending' },
    approved: { color: 'bg-green-100 text-green-800', label: 'Approved' },
    rejected: { color: 'bg-red-100 text-red-800', label: 'Rejected' },
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  const handleReject = async () => {
    if (currentExemption && rejectReason.trim()) {
      await onReject?.(currentExemption.id, rejectReason.trim());
      setShowRejectModal(false);
      setCurrentExemption(null);
      setRejectReason('');
    }
  };

  return (
    <div className="space-y-4">
      {exemptions.map((exemption) => (
        <Card key={exemption.id}>
          <CardBody>
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-medium text-slate-900">{exemption.requirementTitle}</h3>
                <p className="text-sm text-slate-500 mt-1">{exemption.reason}</p>
                <div className="flex items-center gap-4 mt-3 text-sm text-slate-500">
                  <span>Requested: {formatDate(exemption.requestedAt)}</span>
                  <span>By: {exemption.requestedBy}</span>
                  {exemption.expiryDate && (
                    <span>Expires: {formatDate(exemption.expiryDate)}</span>
                  )}
                </div>
                {exemption.reviewNotes && (
                  <div className="mt-3 bg-slate-50 rounded-lg p-3">
                    <p className="text-xs font-medium text-slate-500">Review Notes</p>
                    <p className="text-sm text-slate-700 mt-1">{exemption.reviewNotes}</p>
                  </div>
                )}
              </div>
              <div className="flex items-center gap-2">
                <span
                  className={`px-2 py-1 text-xs font-medium rounded-full ${statusConfig[exemption.status].color}`}
                >
                  {statusConfig[exemption.status].label}
                </span>
              </div>
            </div>

            {exemption.status === 'pending' && onApprove && onReject && (
              <div className="flex justify-end gap-2 mt-4 pt-4 border-t border-slate-200">
                <Button
                  variant="secondary"
                  size="sm"
                  onClick={() => {
                    setCurrentExemption(exemption);
                    setShowRejectModal(true);
                  }}
                  disabled={loading}
                >
                  Reject
                </Button>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => onApprove(exemption.id)}
                  disabled={loading}
                >
                  Approve
                </Button>
              </div>
            )}
          </CardBody>
        </Card>
      ))}

      {exemptions.length === 0 && (
        <div className="text-center py-8 bg-slate-50 rounded-lg">
          <p className="text-slate-500">No exemption requests found</p>
        </div>
      )}

      {/* Reject Modal */}
      <Modal
        isOpen={showRejectModal}
        onClose={() => setShowRejectModal(false)}
        title="Reject Exemption Request"
        size="md"
      >
        <div className="space-y-4">
          <Textarea
            label="Rejection Reason"
            value={rejectReason}
            onChange={(e) => setRejectReason(e.target.value)}
            placeholder="Provide a reason for rejecting this exemption request..."
            rows={4}
            fullWidth
          />
          <div className="flex justify-end gap-3">
            <Button variant="secondary" onClick={() => setShowRejectModal(false)}>
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleReject}
              disabled={!rejectReason.trim()}
            >
              Reject Request
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
};

export default ExemptionForm;
