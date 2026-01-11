/**
 * NEAM Dashboard - Interventions Panel Component
 * Display and manage policy interventions
 */

import React, { useState, useMemo } from 'react';
import { useAppSelector, useAppDispatch } from '../../store';
import { setSelectedIntervention } from '../../store/slices/interventionsSlice';
import {
  Play,
  Pause,
  CheckCircle,
  Clock,
  AlertCircle,
  Filter,
  Search,
  ChevronRight,
  TrendingUp,
  Target,
  Calendar,
} from 'lucide-react';
import { format, formatDistanceToNow } from 'date-fns';
import type { Intervention, InterventionStatus, InterventionType } from '../../types';
import clsx from 'clsx';

// Status configuration
const STATUS_CONFIG: Record<InterventionStatus, { icon: any; color: string; bgColor: string }> = {
  PENDING: { icon: Clock, color: 'text-amber-600', bgColor: 'bg-amber-50' },
  APPROVED: { icon: CheckCircle, color: 'text-blue-600', bgColor: 'bg-blue-50' },
  EXECUTING: { icon: Play, color: 'text-emerald-600', bgColor: 'bg-emerald-50' },
  IN_PROGRESS: { icon: Play, color: 'text-emerald-600', bgColor: 'bg-emerald-50' },
  COMPLETED: { icon: CheckCircle, color: 'text-emerald-600', bgColor: 'bg-emerald-50' },
  CANCELLED: { icon: AlertCircle, color: 'text-red-600', bgColor: 'bg-red-50' },
  FAILED: { icon: AlertCircle, color: 'text-red-600', bgColor: 'bg-red-50' },
  ON_HOLD: { icon: Pause, color: 'text-slate-600', bgColor: 'bg-slate-100' },
};

// Intervention type display names
const INTERVENTION_TYPE_NAMES: Record<InterventionType, string> = {
  FISCAL_STIMULUS: 'Fiscal Stimulus',
  TAX_RELIEF: 'Tax Relief',
  SUBSIDY_PROGRAM: 'Subsidy Program',
  INFRASTRUCTURE_PROJECT: 'Infrastructure',
  EMPLOYMENT_SCHEME: 'Employment Scheme',
  TRAINING_PROGRAM: 'Training Program',
  REGULATORY_REFORM: 'Regulatory Reform',
};

// Progress bar component
interface ProgressBarProps {
  progress: number;
  status?: InterventionStatus;
  size?: 'sm' | 'md' | 'lg';
}

const ProgressBar: React.FC<ProgressBarProps> = ({ progress, status, size = 'md' }) => {
  const getColor = () => {
    if (status === 'COMPLETED') return 'bg-emerald-500';
    if (status === 'FAILED' || status === 'CANCELLED') return 'bg-red-500';
    if (progress >= 75) return 'bg-emerald-500';
    if (progress >= 50) return 'bg-amber-500';
    return 'bg-blue-500';
  };

  const heights = { sm: 'h-1', md: 'h-2', lg: 'h-3' };

  return (
    <div className={clsx('w-full bg-slate-100 rounded-full overflow-hidden', heights[size])}>
      <div
        className={clsx('h-full rounded-full transition-all duration-500', getColor())}
        style={{ width: `${progress}%` }}
      />
    </div>
  );
};

// Intervention Card
interface InterventionCardProps {
  intervention: Intervention;
  onClick?: () => void;
}

const InterventionCard: React.FC<InterventionCardProps> = ({ intervention, onClick }) => {
  const statusConf = STATUS_CONFIG[intervention.status as InterventionStatus] || STATUS_CONFIG.PENDING;
  const StatusIcon = statusConf.icon;
  const typeName = INTERVENTION_TYPE_NAMES[intervention.type as InterventionType] || intervention.type;

  return (
    <div
      className="p-4 bg-white rounded-lg border border-slate-200 hover:border-slate-300 hover:shadow-md transition-all cursor-pointer"
      onClick={onClick}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className={clsx('px-2 py-0.5 rounded-full text-xs font-medium', statusConf.bgColor, statusConf.color)}>
              {intervention.status.replace('_', ' ')}
            </span>
            <span className="text-xs text-slate-500">{typeName}</span>
          </div>
          <h4 className="font-medium text-slate-800 truncate">{intervention.name}</h4>
          <p className="text-sm text-slate-500 mt-1 line-clamp-2">{intervention.description}</p>
        </div>
        <ChevronRight className="w-5 h-5 text-slate-400 flex-shrink-0" />
      </div>

      {/* Progress */}
      <div className="mt-4">
        <div className="flex items-center justify-between mb-1">
          <span className="text-xs text-slate-500">Progress</span>
          <span className="text-xs font-medium text-slate-700">{intervention.progress}%</span>
        </div>
        <ProgressBar progress={intervention.progress} status={intervention.status as InterventionStatus} />
      </div>

      {/* Metadata */}
      <div className="flex items-center gap-4 mt-3 pt-3 border-t border-slate-100">
        <div className="flex items-center gap-1 text-xs text-slate-500">
          <Target className="w-3.5 h-3.5" />
          <span>{intervention.budgetUtilized.toLocaleString()} / {intervention.budgetAllocated.toLocaleString()}</span>
        </div>
        <div className="flex items-center gap-1 text-xs text-slate-500">
          <Calendar className="w-3.5 h-3.5" />
          <span>{formatDistanceToNow(new Date(intervention.triggeredAt), { addSuffix: true })}</span>
        </div>
      </div>
    </div>
  );
};

interface InterventionsPanelProps {
  maxItems?: number;
  showFilters?: boolean;
  compact?: boolean;
}

export const InterventionsPanel: React.FC<InterventionsPanelProps> = ({
  maxItems,
  showFilters = true,
  compact = false,
}) => {
  const dispatch = useAppDispatch();
  const { interventions, filters } = useAppSelector((state) => state.interventions);
  const [searchTerm, setSearchTerm] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');
  const [typeFilter, setTypeFilter] = useState<string>('');

  // Filter interventions
  const filteredInterventions = useMemo(() => {
    let result = [...interventions];

    // Search filter
    if (searchTerm) {
      result = result.filter(
        (int) =>
          int.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          int.description.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Status filter
    if (statusFilter) {
      result = result.filter((int) => int.status === statusFilter);
    }

    // Type filter
    if (typeFilter) {
      result = result.filter((int) => int.type === typeFilter);
    }

    // Sort by status (active first) and date
    const statusOrder = {
      EXECUTING: 0,
      IN_PROGRESS: 1,
      APPROVED: 2,
      PENDING: 3,
      ON_HOLD: 4,
      COMPLETED: 5,
      CANCELLED: 6,
      FAILED: 7,
    };
    result.sort((a, b) => {
      const orderDiff = (statusOrder[a.status as keyof typeof statusOrder] || 10) -
                       (statusOrder[b.status as keyof typeof statusOrder] || 10);
      if (orderDiff !== 0) return orderDiff;
      return new Date(b.triggeredAt).getTime() - new Date(a.triggeredAt).getTime();
    });

    return maxItems ? result.slice(0, maxItems) : result;
  }, [interventions, searchTerm, statusFilter, typeFilter, maxItems]);

  // Stats
  const stats = useMemo(() => ({
    total: interventions.length,
    active: interventions.filter((i) => ['EXECUTING', 'IN_PROGRESS'].includes(i.status)).length,
    completed: interventions.filter((i) => i.status === 'COMPLETED').length,
    pending: interventions.filter((i) => ['PENDING', 'APPROVED'].includes(i.status)).length,
  }), [interventions]);

  const handleInterventionClick = (intervention: Intervention) => {
    dispatch(setSelectedIntervention(intervention));
  };

  if (compact) {
    return (
      <div className="space-y-3">
        {filteredInterventions.slice(0, 3).map((intervention) => (
          <InterventionCard
            key={intervention.id}
            intervention={intervention}
            onClick={() => handleInterventionClick(intervention)}
          />
        ))}
      </div>
    );
  }

  return (
    <div className="bg-white rounded-xl border border-slate-200">
      {/* Header with Stats */}
      <div className="p-4 border-b border-slate-200">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-slate-800">Interventions</h3>
          <div className="flex items-center gap-3 text-xs">
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 bg-emerald-500 rounded-full" />
              <span className="text-slate-600">{stats.active} active</span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 bg-blue-500 rounded-full" />
              <span className="text-slate-600">{stats.pending} pending</span>
            </span>
            <span className="flex items-center gap-1">
              <span className="w-2 h-2 bg-slate-400 rounded-full" />
              <span className="text-slate-600">{stats.completed} done</span>
            </span>
          </div>
        </div>

        {/* Search and Filters */}
        {showFilters && (
          <div className="flex gap-3">
            <div className="relative flex-1">
              <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
              <input
                type="text"
                placeholder="Search interventions..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Status</option>
              <option value="EXECUTING">Executing</option>
              <option value="IN_PROGRESS">In Progress</option>
              <option value="APPROVED">Approved</option>
              <option value="PENDING">Pending</option>
              <option value="ON_HOLD">On Hold</option>
              <option value="COMPLETED">Completed</option>
            </select>
            <select
              value={typeFilter}
              onChange={(e) => setTypeFilter(e.target.value)}
              className="px-3 py-2 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All Types</option>
              {Object.entries(INTERVENTION_TYPE_NAMES).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </div>
        )}
      </div>

      {/* Intervention List */}
      <div className="p-4 space-y-3 max-h-96 overflow-y-auto">
        {filteredInterventions.length > 0 ? (
          filteredInterventions.map((intervention) => (
            <InterventionCard
              key={intervention.id}
              intervention={intervention}
              onClick={() => handleInterventionClick(intervention)}
            />
          ))
        ) : (
          <div className="text-center py-8">
            <Target className="w-12 h-12 text-slate-300 mx-auto mb-3" />
            <p className="text-slate-500">No interventions found</p>
          </div>
        )}
      </div>
    </div>
  );
};

// Mini Interventions Widget
export const MiniInterventionsWidget: React.FC = () => {
  const { interventions } = useAppSelector((state) => state.interventions);
  const activeInterventions = interventions.filter((i) => 
    ['EXECUTING', 'IN_PROGRESS'].includes(i.status)
  ).slice(0, 3);

  return (
    <div className="p-3 bg-white rounded-lg border border-slate-200">
      <div className="flex items-center justify-between mb-2">
        <span className="text-sm font-medium text-slate-700">Active Interventions</span>
        <span className="px-1.5 py-0.5 bg-emerald-100 text-emerald-600 text-xs font-medium rounded">
          {activeInterventions.length}
        </span>
      </div>
      {activeInterventions.length > 0 ? (
        <div className="space-y-2">
          {activeInterventions.map((int) => (
            <div key={int.id} className="flex items-center gap-2">
              <Play className="w-3 h-3 text-emerald-500" />
              <span className="text-xs text-slate-600 truncate flex-1">{int.name}</span>
              <span className="text-xs font-medium text-slate-700">{int.progress}%</span>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-xs text-slate-400">No active interventions</p>
      )}
    </div>
  );
};

export default InterventionsPanel;
