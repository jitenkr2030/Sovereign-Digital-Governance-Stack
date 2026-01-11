import React, { useState } from 'react';
import { useExportDashboardMutation } from '../store/api/dashboardApi';
import { format } from 'date-fns';
import { Download, FileText, FileSpreadsheet, File, X, Check, Loader2 } from 'lucide-react';

interface ExportModalProps {
  onClose: () => void;
}

type ExportFormat = 'pdf' | 'csv' | 'xlsx';

interface ExportSection {
  id: string;
  label: string;
  selected: boolean;
}

export const ExportModal: React.FC<ExportModalProps> = ({ onClose }) => {
  const [format, setFormat] = useState<ExportFormat>('pdf');
  const [dateRange, setDateRange] = useState({
    start: format(new Date(Date.now() - 30 * 24 * 60 * 60 * 1000), 'yyyy-MM-dd'),
    end: format(new Date(), 'yyyy-MM-dd'),
  });
  const [sections, setSections] = useState<ExportSection[]>([
    { id: 'SUMMARY', label: 'Executive Summary', selected: true },
    { id: 'METRICS', label: 'Key Metrics', selected: true },
    { id: 'ALERTS', label: 'Alert History', selected: true },
    { id: 'INTERVENTIONS', label: 'Intervention Records', selected: true },
    { id: 'REGIONAL_DATA', label: 'Regional Analysis', selected: false },
    { id: 'POLICIES', label: 'Active Policies', selected: false },
  ]);
  const [includeCharts, setIncludeCharts] = useState(true);
  const [isExporting, setIsExporting] = useState(false);
  const [exportSuccess, setExportSuccess] = useState(false);
  
  const [exportDashboard] = useExportDashboardMutation();

  const formatOptions: { value: ExportFormat; label: string; icon: React.ReactNode }[] = [
    { value: 'pdf', label: 'PDF Report', icon: <FileText className="h-5 w-5" /> },
    { value: 'csv', label: 'CSV Data', icon: <FileSpreadsheet className="h-5 w-5" /> },
    { value: 'xlsx', label: 'Excel Workbook', icon: <File className="h-5 w-5" /> },
  ];

  const toggleSection = (sectionId: string) => {
    setSections(sections.map(section => 
      section.id === sectionId 
        ? { ...section, selected: !section.selected } 
        : section
    ));
  };

  const handleExport = async () => {
    const selectedSections = sections.filter(s => s.selected).map(s => s.id);
    
    if (selectedSections.length === 0) {
      alert('Please select at least one section to export');
      return;
    }

    setIsExporting(true);
    setExportSuccess(false);

    try {
      const response = await exportDashboard({
        format,
        sections: selectedSections,
        dateRange: {
          start: new Date(dateRange.start),
          end: new Date(dateRange.end),
        },
      });

      // Handle blob response
      if ('data' in response) {
        const blob = response.data as Blob;
        const url = window.URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `neam-report-${format(new Date(), 'yyyy-MM-dd')}.${format}`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        window.URL.revokeObjectURL(url);
        
        setExportSuccess(true);
        setTimeout(() => {
          setExportSuccess(false);
        }, 3000);
      }
    } catch (error) {
      console.error('Export failed:', error);
      alert('Export failed. Please try again.');
    } finally {
      setIsExporting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div 
        className="absolute inset-0 bg-black bg-opacity-50"
        onClick={onClose}
      />
      
      {/* Modal */}
      <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 animate-fade-in">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b">
          <div className="flex items-center space-x-3">
            <Download className="h-5 w-5 text-blue-600" />
            <h2 className="text-lg font-semibold text-gray-900">Export Dashboard Report</h2>
          </div>
          <button 
            onClick={onClose}
            className="p-1 hover:bg-gray-100 rounded"
          >
            <X className="h-5 w-5 text-gray-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-5">
          {/* Format Selection */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Export Format
            </label>
            <div className="grid grid-cols-3 gap-2">
              {formatOptions.map((option) => (
                <button
                  key={option.value}
                  onClick={() => setFormat(option.value)}
                  className={`flex flex-col items-center justify-center p-3 rounded-lg border-2 transition-all ${
                    format === option.value
                      ? 'border-blue-500 bg-blue-50 text-blue-700'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  {option.icon}
                  <span className="mt-1 text-sm font-medium">{option.label}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Date Range */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Date Range
            </label>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-xs text-gray-500 mb-1">Start Date</label>
                <input
                  type="date"
                  value={dateRange.start}
                  onChange={(e) => setDateRange({ ...dateRange, start: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-xs text-gray-500 mb-1">End Date</label>
                <input
                  type="date"
                  value={dateRange.end}
                  onChange={(e) => setDateRange({ ...dateRange, end: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
          </div>

          {/* Sections */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Include Sections
            </label>
            <div className="space-y-2">
              {sections.map((section) => (
                <label 
                  key={section.id}
                  className={`flex items-center p-3 rounded-lg border cursor-pointer transition-all ${
                    section.selected
                      ? 'border-blue-300 bg-blue-50'
                      : 'border-gray-200 hover:bg-gray-50'
                  }`}
                >
                  <input
                    type="checkbox"
                    checked={section.selected}
                    onChange={() => toggleSection(section.id)}
                    className="h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
                  />
                  <span className="ml-3 text-sm text-gray-700">{section.label}</span>
                </label>
              ))}
            </div>
          </div>

          {/* Options */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Options
            </label>
            <div className="space-y-2">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={includeCharts}
                  onChange={(e) => setIncludeCharts(e.target.checked)}
                  className="h-4 w-4 text-blue-600 rounded border-gray-300 focus:ring-blue-500"
                />
                <span className="ml-3 text-sm text-gray-700">Include charts and visualizations</span>
              </label>
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end space-x-3 p-4 border-t bg-gray-50 rounded-b-xl">
          <button
            onClick={onClose}
            className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-900"
          >
            Cancel
          </button>
          <button
            onClick={handleExport}
            disabled={isExporting || exportSuccess}
            className={`flex items-center space-x-2 px-4 py-2 rounded-lg text-sm font-medium text-white transition-all ${
              exportSuccess
                ? 'bg-green-600'
                : 'bg-blue-600 hover:bg-blue-700'
            } disabled:opacity-50 disabled:cursor-not-allowed`}
          >
            {isExporting ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Exporting...</span>
              </>
            ) : exportSuccess ? (
              <>
                <Check className="h-4 w-4" />
                <span>Downloaded!</span>
              </>
            ) : (
              <>
                <Download className="h-4 w-4" />
                <span>Export Report</span>
              </>
            )}
          </button>
        </div>
      </div>
    </div>
  );
};
