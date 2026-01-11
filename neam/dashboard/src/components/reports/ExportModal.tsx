/**
 * NEAM Dashboard - Export Modal Component
 * Generate PDF and CSV exports for reports
 */

import React, { useState } from 'react';
import { useAppSelector, useAppDispatch } from '../../store';
import { closeModal } from '../../store/slices/uiSlice';
import {
  X,
  FileText,
  FileSpreadsheet,
  Download,
  Calendar,
  Layers,
  MapPin,
  BarChart3,
  Bell,
  Target,
  Check,
  Loader2,
} from 'lucide-react';
import { format } from 'date-fns';
import type { ExportConfig, ExportSection, ReportType } from '../../types';
import clsx from 'clsx';

// Section options with icons
const SECTION_OPTIONS: { id: ExportSection; label: string; icon: any; description: string }[] = [
  { id: 'SUMMARY', label: 'Executive Summary', icon: FileText, description: 'Key metrics overview' },
  { id: 'METRICS', label: 'Detailed Metrics', icon: BarChart3, description: 'All economic indicators' },
  { id: 'REGIONAL_DATA', label: 'Regional Analysis', icon: MapPin, description: 'State and district data' },
  { id: 'ALERTS', label: 'Alert Log', icon: Bell, description: 'Active and historical alerts' },
  { id: 'INTERVENTIONS', label: 'Interventions', icon: Target, description: 'Policy intervention status' },
  { id: 'ANALYTICS', label: 'Analytics', icon: BarChart3, description: 'Trend analysis' },
];

// Report types
const REPORT_TYPES: { id: ReportType; label: string; description: string }[] = [
  { id: 'PERFORMANCE_SUMMARY', description: 'Overview of economic performance' },
  { id: 'REGIONAL_ANALYSIS', description: 'Detailed regional comparison' },
  { id: 'SECTOR_BREAKDOWN', description: 'Sector-wise analysis' },
  { id: 'ALERT_LOG', description: 'Summary of alerts and incidents' },
  { id: 'COMPARATIVE_STUDY', description: 'Comparison across regions/periods' },
  { id: 'TREND_REPORT', description: 'Historical trends and forecasts' },
  { id: 'AUDIT_TRAIL', description: 'System activity log' },
];

// Confidentiality levels
const CONFIDENTIALITY_LEVELS = [
  { id: 'public', label: 'Public', color: 'bg-green-100 text-green-700' },
  { id: 'restricted', label: 'Restricted', color: 'bg-blue-100 text-blue-700' },
  { id: 'confidential', label: 'Confidential', color: 'bg-amber-100 text-amber-700' },
  { id: 'secret', label: 'Secret', color: 'bg-red-100 text-red-700' },
];

interface ExportModalProps {
  onExport?: (config: ExportConfig) => Promise<void>;
}

export const ExportModal: React.FC<ExportModalProps> = ({ onExport }) => {
  const dispatch = useAppDispatch();
  const { dateRange } = useAppSelector((state) => state.filters);
  const [isExporting, setIsExporting] = useState(false);
  const [activeTab, setActiveTab] = useState<'quick' | 'custom'>('quick');
  
  // Form state
  const [reportType, setReportType] = useState<ReportType>('PERFORMANCE_SUMMARY');
  const [format, setFormat] = useState<'pdf' | 'csv' | 'xlsx'>('pdf');
  const [sections, setSections] = useState<ExportSection[]>(['SUMMARY', 'METRICS']);
  const [selectedRegions, setSelectedRegions] = useState<string[]>([]);
  const [includeCharts, setIncludeCharts] = useState(true);
  const [includeTables, setIncludeTables] = useState(true);
  const [confidentiality, setConfidentiality] = useState('restricted');
  const [customTitle, setCustomTitle] = useState('');

  const handleClose = () => {
    dispatch(closeModal('exportModal'));
  };

  const handleSectionToggle = (sectionId: ExportSection) => {
    setSections((prev) =>
      prev.includes(sectionId)
        ? prev.filter((s) => s !== sectionId)
        : [...prev, sectionId]
    );
  };

  const handleExport = async () => {
    setIsExporting(true);
    try {
      const config: ExportConfig = {
        type: reportType,
        format,
        dateRange: {
          start: dateRange.start,
          end: dateRange.end,
        },
        sections,
        includeCharts,
        includeTables,
        confidentialityLevel: confidentiality,
        title: customTitle || REPORT_TYPES.find((r) => r.id === reportType)?.description,
      };
      
      if (onExport) {
        await onExport(config);
      } else {
        // Default export behavior
        console.log('Export config:', config);
        // Simulate export delay
        await new Promise((resolve) => setTimeout(resolve, 2000));
      }
      
      handleClose();
    } catch (error) {
      console.error('Export failed:', error);
    } finally {
      setIsExporting(false);
    }
  };

  const isValid = sections.length > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/50 backdrop-blur-sm"
        onClick={handleClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[90vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
          <div>
            <h2 className="text-xl font-semibold text-slate-800">Export Report</h2>
            <p className="text-sm text-slate-500 mt-0.5">
              Generate PDF, CSV, or Excel reports for Parliament & courts
            </p>
          </div>
          <button
            onClick={handleClose}
            className="p-2 rounded-lg hover:bg-slate-100 transition-colors"
          >
            <X className="w-5 h-5 text-slate-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          {/* Date Range */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              <Calendar className="w-4 h-4 inline mr-1" />
              Date Range
            </label>
            <div className="flex items-center gap-3">
              <input
                type="date"
                value={format(dateRange.start, 'yyyy-MM-dd')}
                onChange={(e) => {}}
                className="px-3 py-2 border border-slate-200 rounded-lg text-sm bg-slate-50"
                disabled
              />
              <span className="text-slate-400">to</span>
              <input
                type="date"
                value={format(dateRange.end, 'yyyy-MM-dd')}
                onChange={(e) => {}}
                className="px-3 py-2 border border-slate-200 rounded-lg text-sm bg-slate-50"
                disabled
              />
            </div>
          </div>

          {/* Format Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Export Format
            </label>
            <div className="flex gap-3">
              {(['pdf', 'csv', 'xlsx'] as const).map((f) => {
                const Icon = f === 'pdf' ? FileText : FileSpreadsheet;
                return (
                  <button
                    key={f}
                    onClick={() => setFormat(f)}
                    className={clsx(
                      'flex items-center gap-2 px-4 py-3 rounded-lg border transition-all',
                      format === f
                        ? 'border-blue-500 bg-blue-50 text-blue-700'
                        : 'border-slate-200 hover:border-slate-300'
                    )}
                  >
                    <Icon className="w-5 h-5" />
                    <span className="font-medium uppercase">{f}</span>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Report Type */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Report Type
            </label>
            <select
              value={reportType}
              onChange={(e) => setReportType(e.target.value as ReportType)}
              className="w-full px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {REPORT_TYPES.map((type) => (
                <option key={type.id} value={type.id}>
                  {type.description}
                </option>
              ))}
            </select>
          </div>

          {/* Sections */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              <Layers className="w-4 h-4 inline mr-1" />
              Include Sections
            </label>
            <div className="grid grid-cols-2 gap-3">
              {SECTION_OPTIONS.map((section) => {
                const Icon = section.icon;
                const isSelected = sections.includes(section.id);
                return (
                  <button
                    key={section.id}
                    onClick={() => handleSectionToggle(section.id)}
                    className={clsx(
                      'flex items-start gap-3 p-3 rounded-lg border text-left transition-all',
                      isSelected
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-slate-200 hover:border-slate-300'
                    )}
                  >
                    <div className={clsx(
                      'p-1.5 rounded-lg',
                      isSelected ? 'bg-blue-100' : 'bg-slate-100'
                    )}>
                      <Icon className={clsx(
                        'w-4 h-4',
                        isSelected ? 'text-blue-600' : 'text-slate-500'
                      )} />
                    </div>
                    <div>
                      <p className={clsx(
                        'text-sm font-medium',
                        isSelected ? 'text-blue-700' : 'text-slate-700'
                      )}>
                        {section.label}
                      </p>
                      <p className="text-xs text-slate-500">{section.description}</p>
                    </div>
                    {isSelected && (
                      <Check className="w-4 h-4 text-blue-600 ml-auto" />
                    )}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Options */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Options
            </label>
            <div className="flex flex-wrap gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeCharts}
                  onChange={(e) => setIncludeCharts(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm text-slate-600">Include Charts</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="checkbox"
                  checked={includeTables}
                  onChange={(e) => setIncludeTables(e.target.checked)}
                  className="w-4 h-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-sm text-slate-600">Include Tables</span>
              </label>
            </div>
          </div>

          {/* Confidentiality */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-slate-700 mb-2">
              Confidentiality Level
            </label>
            <div className="flex gap-2">
              {CONFIDENTIALITY_LEVELS.map((level) => (
                <button
                  key={level.id}
                  onClick={() => setConfidentiality(level.id)}
                  className={clsx(
                    'px-3 py-1.5 rounded-lg text-sm font-medium transition-all',
                    confidentiality === level.id
                      ? level.color
                      : 'bg-slate-100 text-slate-600 hover:bg-slate-200'
                  )}
                >
                  {level.label}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-6 py-4 border-t border-slate-200 bg-slate-50">
          <div className="text-sm text-slate-500">
            {sections.length} sections selected
          </div>
          <div className="flex gap-3">
            <button
              onClick={handleClose}
              className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-800 transition-colors"
              disabled={isExporting}
            >
              Cancel
            </button>
            <button
              onClick={handleExport}
              disabled={!isValid || isExporting}
              className={clsx(
                'flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-all',
                isValid && !isExporting
                  ? 'bg-blue-600 hover:bg-blue-700'
                  : 'bg-slate-300 cursor-not-allowed'
              )}
            >
              {isExporting ? (
                <>
                  <Loader2 className="w-4 h-4 animate-spin" />
                  Exporting...
                </>
              ) : (
                <>
                  <Download className="w-4 h-4" />
                  Export Report
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

// Export button that opens the modal
export const ExportButton: React.FC<{ onClick?: () => void }> = ({ onClick }) => {
  const dispatch = useAppDispatch();

  const handleClick = () => {
    dispatch(closeModal('exportModal'));
    onClick?.();
  };

  return (
    <button
      onClick={handleClick}
      className="flex items-center gap-2 px-4 py-2 bg-slate-800 text-white rounded-lg hover:bg-slate-700 transition-colors"
    >
      <Download className="w-4 h-4" />
      <span className="text-sm font-medium">Export</span>
    </button>
  );
};

export default ExportModal;
