/**
 * NEAM Dashboard - Interventions Page
 * Manage and monitor policy interventions
 */

import React, { useState } from 'react';
import { useAppSelector, useAppDispatch } from '../../store';
import { setSelectedIntervention } from '../../store/slices/interventionsSlice';
import { InterventionsPanel } from '../../components/dashboard/InterventionsPanel';
import { Section, ExportButton } from '../../components/layout/Layout';
import { Play, Pause, Plus, Filter, Search, Target, Calendar, TrendingUp } from 'lucide-react';
import type { Intervention } from '../../types';
import clsx from 'clsx';

// Mock data
const mockInterventions: Intervention[] = [
  {
    id: '1',
    policyId: 'pol-1',
    policyName: 'MSME Support Scheme',
    workflowId: 'wf-1',
    status: 'EXECUTING',
    regionId: 'mh',
    sectorId: 'manufacturing',
    triggeredAt: new Date('2024-01-15'),
    progress: 65,
    currentStep: 'Disbursement',
    type: 'SUBSIDY_PROGRAM',
    name: 'MSME Support Scheme',
    description: 'Financial support package for micro, small, and medium enterprises',
    budgetAllocated: 50000,
    budgetUtilized: 32500,
    targetRegions: ['mh', 'gj', 'ka'],
    targetSectors: ['manufacturing', 'services'],
    expectedOutcome: 'Support 1 lakh MSMEs',
    kpis: [
      { id: 'k1', name: 'MSMEs Supported', target: 100000, current: 65000, unit: 'units', progress: 65 },
      { id: 'k2', name: 'Jobs Created', target: 500000, current: 280000, unit: 'jobs', progress: 56 },
    ],
  },
  {
    id: '2',
    policyId: 'pol-2',
    policyName: 'Agricultural Reform',
    workflowId: 'wf-2',
    status: 'IN_PROGRESS',
    regionId: 'up',
    sectorId: 'agriculture',
    triggeredAt: new Date('2024-02-01'),
    progress: 35,
    currentStep: 'Implementation',
    type: 'REGULATORY_REFORM',
    name: 'Agricultural Marketing Reform',
    description: 'Reform of agricultural marketing regulations to improve farmer incomes',
    budgetAllocated: 25000,
    budgetUtilized: 8750,
    targetRegions: ['up', 'br', 'pb'],
    targetSectors: ['agriculture'],
    expectedOutcome: 'Increase farmer incomes by 20%',
    kpis: [
      { id: 'k3', name: 'Farmers Enrolled', target: 1000000, current: 350000, unit: 'farmers', progress: 35 },
    ],
  },
  {
    id: '3',
    policyId: 'pol-3',
    policyName: 'Skill Development',
    workflowId: 'wf-3',
    status: 'APPROVED',
    regionId: 'all',
    sectorId: 'education',
    triggeredAt: new Date('2024-03-01'),
    progress: 0,
    currentStep: 'Ready to Launch',
    type: 'TRAINING_PROGRAM',
    name: 'National Skill Development Program',
    description: 'Comprehensive skill development for youth across sectors',
    budgetAllocated: 75000,
    budgetUtilized: 0,
    targetRegions: ['all'],
    targetSectors: ['manufacturing', 'services', 'technology'],
    expectedOutcome: 'Train 50 lakh youth',
    kpis: [
      { id: 'k4', name: 'Youth Trained', target: 5000000, current: 0, unit: 'youth', progress: 0 },
    ],
  },
];

const InterventionsPage: React.FC = () => {
  const dispatch = useAppDispatch();
  const { selectedIntervention } = useAppSelector((state) => state.interventions);
  const [view, setView] = useState<'list' | 'detail'>('list');
  const [searchTerm, setSearchTerm] = useState('');

  const stats = {
    total: mockInterventions.length,
    active: mockInterventions.filter((i) => ['EXECUTING', 'IN_PROGRESS'].includes(i.status)).length,
    completed: mockInterventions.filter((i) => i.status === 'COMPLETED').length,
    totalBudget: mockInterventions.reduce((sum, i) => sum + i.budgetAllocated, 0),
    utilizedBudget: mockInterventions.reduce((sum, i) => sum + i.budgetUtilized, 0),
  };

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-slate-800">Interventions</h1>
          <p className="text-slate-500 mt-1">
            Monitor and manage policy interventions across regions
          </p>
        </div>
        <div className="flex items-center gap-3">
          <button className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors">
            <Plus className="w-4 h-4" />
            <span className="text-sm font-medium">New Intervention</span>
          </button>
          <ExportButton />
        </div>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 rounded-lg">
              <Target className="w-5 h-5 text-blue-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">Total Interventions</p>
              <p className="text-2xl font-bold text-slate-800">{stats.total}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-emerald-100 rounded-lg">
              <Play className="w-5 h-5 text-emerald-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">Active</p>
              <p className="text-2xl font-bold text-slate-800">{stats.active}</p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-amber-100 rounded-lg">
              <TrendingUp className="w-5 h-5 text-amber-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">Avg Progress</p>
              <p className="text-2xl font-bold text-slate-800">
                {Math.round(mockInterventions.reduce((sum, i) => sum + i.progress, 0) / stats.total)}%
              </p>
            </div>
          </div>
        </div>
        <div className="bg-white rounded-xl border border-slate-200 p-4">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 rounded-lg">
              <Calendar className="w-5 h-5 text-purple-600" />
            </div>
            <div>
              <p className="text-sm text-slate-500">Budget Utilization</p>
              <p className="text-2xl font-bold text-slate-800">
                {Math.round((stats.utilizedBudget / stats.totalBudget) * 100)}%
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Interventions List */}
        <div className="lg:col-span-2">
          <Section
            title="All Interventions"
            action={
              <div className="flex items-center gap-2">
                <div className="relative">
                  <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
                  <input
                    type="text"
                    placeholder="Search..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    className="pl-9 pr-4 py-1.5 border border-slate-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <button className="p-2 border border-slate-200 rounded-lg hover:bg-slate-50">
                  <Filter className="w-4 h-4 text-slate-600" />
                </button>
              </div>
            }
          >
            <InterventionsPanel maxItems={10} showFilters={false} />
          </Section>
        </div>

        {/* Selected Intervention Detail */}
        <div className="lg:col-span-1">
          {selectedIntervention ? (
            <Section title="Intervention Details" subtitle={selectedIntervention.name}>
              <div className="space-y-4">
                <div>
                  <p className="text-sm text-slate-500">Status</p>
                  <span className={clsx(
                    'px-2 py-1 rounded-full text-xs font-medium',
                    selectedIntervention.status === 'EXECUTING' ? 'bg-emerald-100 text-emerald-700' :
                    selectedIntervention.status === 'IN_PROGRESS' ? 'bg-blue-100 text-blue-700' :
                    selectedIntervention.status === 'APPROVED' ? 'bg-amber-100 text-amber-700' :
                    'bg-slate-100 text-slate-700'
                  )}>
                    {selectedIntervention.status}
                  </span>
                </div>
                <div>
                  <p className="text-sm text-slate-500">Progress</p>
                  <div className="mt-2">
                    <div className="flex justify-between text-sm mb-1">
                      <span className="text-slate-700">{selectedIntervention.progress}%</span>
                    </div>
                    <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-blue-500 rounded-full transition-all"
                        style={{ width: `${selectedIntervention.progress}%` }}
                      />
                    </div>
                  </div>
                </div>
                <div>
                  <p className="text-sm text-slate-500">Budget</p>
                  <p className="text-lg font-semibold text-slate-800">
                    ₹{selectedIntervention.budgetUtilized.toLocaleString()} / ₹{selectedIntervention.budgetAllocated.toLocaleString()} Cr
                  </p>
                </div>
                <div>
                  <p className="text-sm text-slate-500">Description</p>
                  <p className="text-sm text-slate-700 mt-1">{selectedIntervention.description}</p>
                </div>
                <div>
                  <p className="text-sm text-slate-500 mb-2">KPIs</p>
                  {selectedIntervention.kpis.map((kpi) => (
                    <div key={kpi.id} className="mb-3">
                      <div className="flex justify-between text-sm mb-1">
                        <span className="text-slate-700">{kpi.name}</span>
                        <span className="text-slate-500">
                          {kpi.current.toLocaleString()} / {kpi.target.toLocaleString()} {kpi.unit}
                        </span>
                      </div>
                      <div className="h-1.5 bg-slate-100 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-emerald-500 rounded-full"
                          style={{ width: `${kpi.progress}%` }}
                        />
                      </div>
                    </div>
                  ))}
                </div>
                <div className="flex gap-2 pt-2">
                  <button className="flex-1 px-3 py-2 bg-blue-600 text-white rounded-lg text-sm font-medium hover:bg-blue-700">
                    View Details
                  </button>
                  <button className="px-3 py-2 border border-slate-200 rounded-lg text-sm font-medium hover:bg-slate-50">
                    Edit
                  </button>
                </div>
              </div>
            </Section>
          ) : (
            <div className="bg-white rounded-xl border border-slate-200 p-6 text-center">
              <Target className="w-12 h-12 text-slate-300 mx-auto mb-3" />
              <p className="text-slate-500">Select an intervention to view details</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default InterventionsPage;
