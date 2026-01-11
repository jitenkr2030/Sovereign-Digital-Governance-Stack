import React, { useState, useEffect, useCallback } from 'react';
import { useNotification } from '../../context/NotificationContext';
import { useDataFetch } from '../../hooks';
import {
  ComplianceStatusCard,
  ChecklistViewer,
  ViolationTable,
  ExemptionForm,
  ExemptionList,
  ComplianceReportGenerator,
  AuditTrailViewer,
} from '../../components/compliance';
import {
  Card,
  CardBody,
  CardHeader,
  CardTitle,
  Button,
  Select,
  Modal,
  StatsCard,
  Spinner,
} from '../../components/common';
import {
  ComplianceService,
  ComplianceStatus,
  ComplianceChecklist,
  Violation,
  ExemptionRequest,
  ComplianceReport,
} from '../../services/compliance';

type TabType = 'overview' | 'checklist' | 'violations' | 'exemptions' | 'reports' | 'audit';

const Compliance: React.FC = () => {
  const { showNotification } = useNotification();
  const [activeTab, setActiveTab] = useState<TabType>('overview');
  const [selectedCategory, setSelectedCategory] = useState<string>('');
  const [showExemptionForm, setShowExemptionForm] = useState(false);
  const [selectedRequirement, setSelectedRequirement] = useState<ComplianceChecklist | null>(null);

  // Data state
  const [complianceStatus, setComplianceStatus] = useState<ComplianceStatus | null>(null);
  const [checklist, setChecklist] = useState<ComplianceChecklist[]>([]);
  const [violations, setViolations] = useState<Violation[]>([]);
  const [violationsPage, setViolationsPage] = useState(1);
  const [violationsTotal, setViolationsTotal] = useState(0);
  const [exemptions, setExemptions] = useState<ExemptionRequest[]>([]);
  const [recentReports, setRecentReports] = useState<ComplianceReport[]>([]);
  const [auditEntries, setAuditEntries] = useState<any[]>([]);

  // Fetch compliance status
  const fetchStatus = useCallback(async () => {
    try {
      const status = await ComplianceService.getStatus();
      setComplianceStatus(status);
    } catch (error) {
      showNotification('Failed to fetch compliance status', 'error');
    }
  }, [showNotification]);

  // Fetch checklist
  const fetchChecklist = useCallback(async () => {
    try {
      const items = await ComplianceService.getChecklist(selectedCategory || undefined);
      setChecklist(items);
    } catch (error) {
      showNotification('Failed to fetch compliance checklist', 'error');
    }
  }, [selectedCategory, showNotification]);

  // Fetch violations
  const fetchViolations = useCallback(async () => {
    try {
      const result = await ComplianceService.getViolations({
        page: violationsPage,
        pageSize: 10,
      });
      setViolations(result.items);
      setViolationsTotal(result.total);
    } catch (error) {
      showNotification('Failed to fetch violations', 'error');
    }
  }, [violationsPage, showNotification]);

  // Fetch exemptions
  const fetchExemptions = useCallback(async () => {
    try {
      const items = await ComplianceService.getExemptions();
      setExemptions(items);
    } catch (error) {
      showNotification('Failed to fetch exemptions', 'error');
    }
  }, [showNotification]);

  // Fetch recent reports
  const fetchReports = useCallback(async () => {
    try {
      const result = await ComplianceService.getReports({ pageSize: 10 });
      setRecentReports(result.items);
    } catch (error) {
      showNotification('Failed to fetch reports', 'error');
    }
  }, [showNotification]);

  // Fetch audit trail
  const fetchAuditTrail = useCallback(async (filter?: any) => {
    try {
      const result = await ComplianceService.getAuditTrail(filter || {});
      setAuditEntries(result.items);
    } catch (error) {
      showNotification('Failed to fetch audit trail', 'error');
    }
  }, [showNotification]);

  // Initial data fetch
  useEffect(() => {
    fetchStatus();
    fetchChecklist();
    fetchViolations();
    fetchExemptions();
    fetchReports();
    fetchAuditTrail();
  }, []);

  // Refetch when tab changes
  useEffect(() => {
    if (activeTab === 'violations') {
      fetchViolations();
    } else if (activeTab === 'exemptions') {
      fetchExemptions();
    } else if (activeTab === 'reports') {
      fetchReports();
    } else if (activeTab === 'audit') {
      fetchAuditTrail();
    }
  }, [activeTab]);

  // Handlers
  const handleUpdateChecklistItem = async (
    itemId: string,
    status: ComplianceChecklist['status'],
    notes?: string
  ) => {
    try {
      await ComplianceService.updateChecklistItem(itemId, status, notes);
      showNotification('Checklist item updated successfully', 'success');
      fetchChecklist();
      fetchStatus();
    } catch (error) {
      showNotification('Failed to update checklist item', 'error');
    }
  };

  const handleUpdateViolation = async (
    violationId: string,
    updates: Partial<Violation>
  ) => {
    try {
      await ComplianceService.updateViolation(violationId, updates);
      showNotification('Violation updated successfully', 'success');
      fetchViolations();
    } catch (error) {
      showNotification('Failed to update violation', 'error');
    }
  };

  const handleRequestExemption = async (payload: any) => {
    try {
      await ComplianceService.requestExemption(payload);
      showNotification('Exemption request submitted successfully', 'success');
      setShowExemptionForm(false);
      setSelectedRequirement(null);
      fetchExemptions();
    } catch (error) {
      showNotification('Failed to submit exemption request', 'error');
    }
  };

  const handleApproveExemption = async (exemptionId: string) => {
    try {
      await ComplianceService.approveExemption(exemptionId, 'Approved');
      showNotification('Exemption approved successfully', 'success');
      fetchExemptions();
    } catch (error) {
      showNotification('Failed to approve exemption', 'error');
    }
  };

  const handleRejectExemption = async (exemptionId: string, reason: string) => {
    try {
      await ComplianceService.rejectExemption(exemptionId, reason);
      showNotification('Exemption rejected', 'success');
      fetchExemptions();
    } catch (error) {
      showNotification('Failed to reject exemption', 'error');
    }
  };

  const handleGenerateReport = async (
    type: ComplianceReport['type'],
    period: { start: string; end: string },
    categoryId?: string
  ) => {
    try {
      await ComplianceService.generateReport(type, period, categoryId);
      showNotification('Report generated successfully', 'success');
      fetchReports();
    } catch (error) {
      showNotification('Failed to generate report', 'error');
    }
  };

  const handleDownloadReport = async (reportId: string) => {
    try {
      const blob = await ComplianceService.downloadReport(reportId);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `compliance-report-${reportId}.pdf`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      showNotification('Failed to download report', 'error');
    }
  };

  const handleExportAuditTrail = async (format: 'csv' | 'pdf' | 'json') => {
    try {
      const blob = await ComplianceService.exportAuditTrail({}, format);
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit-trail.${format}`;
      a.click();
      window.URL.revokeObjectURL(url);
    } catch (error) {
      showNotification('Failed to export audit trail', 'error');
    }
  };

  const tabs = [
    { id: 'overview', label: 'Overview', icon: 'M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6' },
    { id: 'checklist', label: 'Checklist', icon: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4' },
    { id: 'violations', label: 'Violations', icon: 'M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z' },
    { id: 'exemptions', label: 'Exemptions', icon: 'M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z' },
    { id: 'reports', label: 'Reports', icon: 'M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
    { id: 'audit', label: 'Audit Trail', icon: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' },
  ];

  const stats = [
    {
      title: 'Overall Score',
      value: `${complianceStatus?.overallScore || 0}%`,
      change: '+2.5%',
      trend: 'up' as const,
      color: complianceStatus?.overallScore >= 80 ? 'green' : complianceStatus?.overallScore >= 60 ? 'amber' : 'red',
    },
    {
      title: 'Open Violations',
      value: violations.filter((v) => v.status === 'open').length.toString(),
      change: '-3',
      trend: 'down' as const,
      color: 'red',
    },
    {
      title: 'Pending Exemptions',
      value: exemptions.filter((e) => e.status === 'pending').length.toString(),
      change: '+1',
      trend: 'up' as const,
      color: 'amber',
    },
    {
      title: 'Compliant Items',
      value: checklist.filter((c) => c.status === 'compliant').length.toString(),
      change: '+5',
      trend: 'up' as const,
      color: 'green',
    },
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-900">Compliance Management</h1>
          <p className="text-slate-500 mt-1">Monitor and manage regulatory compliance</p>
        </div>
        <div className="flex items-center gap-2">
          <Select
            value={selectedCategory}
            onChange={(e) => setSelectedCategory(e.target.value)}
            options={[
              { value: '', label: 'All Categories' },
              { value: 'safety', label: 'Safety' },
              { value: 'environmental', label: 'Environmental' },
              { value: 'operational', label: 'Operational' },
              { value: 'financial', label: 'Financial' },
            ]}
          />
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <StatsCard key={stat.title} {...stat} />
        ))}
      </div>

      {/* Tabs */}
      <div className="border-b border-slate-200">
        <nav className="flex space-x-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as TabType)}
              className={`
                flex items-center gap-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors
                ${activeTab === tab.id
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-slate-500 hover:text-slate-700 hover:border-slate-300'
                }
              `}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d={tab.icon} />
              </svg>
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="min-h-[400px]">
        {activeTab === 'overview' && (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {complianceStatus && (
              <ComplianceStatusCard status={complianceStatus} onRefresh={fetchStatus} />
            )}
            <Card>
              <CardHeader>
                <CardTitle>Recent Activity</CardTitle>
              </CardHeader>
              <CardBody>
                <div className="space-y-4">
                  {auditEntries.slice(0, 5).map((entry) => (
                    <div key={entry.id} className="flex items-start gap-3 pb-3 border-b border-slate-100 last:border-0">
                      <div className="p-2 bg-blue-100 rounded-lg">
                        <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                        </svg>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-slate-900">{entry.action}</p>
                        <p className="text-sm text-slate-500">{entry.entityName}</p>
                        <p className="text-xs text-slate-400 mt-1">{new Date(entry.performedAt).toLocaleString()}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </CardBody>
            </Card>
          </div>
        )}

        {activeTab === 'checklist' && (
          <ChecklistViewer
            checklist={checklist}
            onUpdateItem={handleUpdateChecklistItem}
            onViewEvidence={(item) => console.log('View evidence:', item)}
          />
        )}

        {activeTab === 'violations' && (
          <ViolationTable
            violations={violations}
            total={violationsTotal}
            currentPage={violationsPage}
            pageSize={10}
            onPageChange={setViolationsPage}
            onUpdateViolation={handleUpdateViolation}
            onViewDetails={(violation) => console.log('View details:', violation)}
          />
        )}

        {activeTab === 'exemptions' && (
          <ExemptionList
            exemptions={exemptions}
            onApprove={handleApproveExemption}
            onReject={handleRejectExemption}
          />
        )}

        {activeTab === 'reports' && (
          <ComplianceReportGenerator
            onGenerate={handleGenerateReport}
            onDownload={handleDownloadReport}
            onView={(report) => console.log('View report:', report)}
            recentReports={recentReports}
          />
        )}

        {activeTab === 'audit' && (
          <AuditTrailViewer
            entries={auditEntries}
            onExport={handleExportAuditTrail}
            onFilter={fetchAuditTrail}
          />
        )}
      </div>

      {/* Exemption Form Modal */}
      <Modal
        isOpen={showExemptionForm}
        onClose={() => {
          setShowExemptionForm(false);
          setSelectedRequirement(null);
        }}
        title="Request Exemption"
        size="lg"
      >
        {selectedRequirement && (
          <ExemptionForm
            requirement={selectedRequirement}
            onSubmit={handleRequestExemption}
            onClose={() => {
              setShowExemptionForm(false);
              setSelectedRequirement(null);
            }}
          />
        )}
      </Modal>
    </div>
  );
};

export default Compliance;
