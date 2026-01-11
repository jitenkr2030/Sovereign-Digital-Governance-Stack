import React, { useState, useCallback } from 'react';
import { useSelector, useDispatch } from 'react-redux';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import { useGetDashboardSummaryQuery, useGetMetricCardsQuery, useGetRegionsQuery } from '../store/api/dashboardApi';
import { setViewMode } from '../store/slices/dashboardSlice';
import type { RootState } from '../store';
import { MetricCard } from './MetricCard';
import { RegionHeatmap } from './RegionHeatmap';
import { AlertsPanel } from './AlertsPanel';
import { InterventionsPanel } from './InterventionsPanel';
import { ExportModal } from './ExportModal';
import { useNavigate } from 'react-router-dom';
import { 
  LayoutDashboard, 
  Map, 
  Bell, 
  Settings, 
  User, 
  Download,
  RefreshCw,
  AlertTriangle,
  Activity,
  TrendingUp
} from 'lucide-react';

export const Dashboard: React.FC = () => {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { user } = useSelector((state: RootState) => state.auth);
  const { viewMode } = useSelector((state: RootState) => state.dashboard);
  
  const { data: summary, isLoading: summaryLoading, refetch: refetchSummary } = useGetDashboardSummaryQuery();
  const { data: metrics, isLoading: metricsLoading } = useGetMetricCardsQuery();
  const { data: regions } = useGetRegionsQuery({});
  
  const [showExport, setShowExport] = useState(false);
  const [activePanel, setActivePanel] = useState<'alerts' | 'interventions' | null>(null);

  const handleViewModeChange = useCallback((mode: 'overview' | 'regional' | 'district') => {
    dispatch(setViewMode(mode));
  }, [dispatch]);

  const getSystemHealthColor = (health: string | undefined) => {
    switch (health) {
      case 'healthy': return 'text-green-500';
      case 'warning': return 'text-yellow-500';
      case 'critical': return 'text-red-500';
      default: return 'text-gray-500';
    }
  };

  const getSystemHealthBg = (health: string | undefined) => {
    switch (health) {
      case 'healthy': return 'bg-green-100 border-green-300';
      case 'warning': return 'bg-yellow-100 border-yellow-300';
      case 'critical': return 'bg-red-100 border-red-300';
      default: return 'bg-gray-100 border-gray-300';
    }
  };

  if (summaryLoading || metricsLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200 sticky top-0 z-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center h-16">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <Activity className="h-8 w-8 text-blue-600" />
                <span className="text-xl font-bold text-gray-900">NEAM</span>
              </div>
              <span className="text-sm text-gray-500">National Economic Activity Monitor</span>
            </div>
            
            <div className="flex items-center space-x-4">
              {/* System Health Indicator */}
              <div className={`px-3 py-1 rounded-full border flex items-center space-x-2 ${getSystemHealthBg(summary?.systemHealth)}`}>
                <span className={`w-2 h-2 rounded-full ${
                  summary?.systemHealth === 'healthy' ? 'bg-green-500' :
                  summary?.systemHealth === 'warning' ? 'bg-yellow-500' : 'bg-red-500'
                }`}></span>
                <span className={`text-sm font-medium ${getSystemHealthColor(summary?.systemHealth)}`}>
                  System {summary?.systemHealth || 'Unknown'}
                </span>
              </div>
              
              <button 
                onClick={() => refetchSummary()}
                className="p-2 text-gray-500 hover:text-gray-700 hover:bg-gray-100 rounded-full"
              >
                <RefreshCw className="h-5 w-5" />
              </button>
              
              <button 
                onClick={() => setShowExport(true)}
                className="flex items-center space-x-2 px-3 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
              >
                <Download className="h-4 w-4" />
                <span>Export</span>
              </button>
              
              <div className="flex items-center space-x-2">
                <User className="h-8 w-8 text-gray-400" />
                <div className="text-sm">
                  <p className="font-medium text-gray-900">{user?.name || 'Guest'}</p>
                  <p className="text-gray-500">{user?.role || 'No Role'}</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {/* KPI Ticker */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">Active Alerts</p>
                <p className="text-2xl font-bold text-gray-900">{summary?.activeAlerts || 0}</p>
              </div>
              <AlertTriangle className="h-8 w-8 text-orange-500" />
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">Active Interventions</p>
                <p className="text-2xl font-bold text-gray-900">{summary?.activeInterventions || 0}</p>
              </div>
              <TrendingUp className="h-8 w-8 text-blue-500" />
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">Regions Monitored</p>
                <p className="text-2xl font-bold text-gray-900">{summary?.regionsMonitored || 0}</p>
              </div>
              <Map className="h-8 w-8 text-green-500" />
            </div>
          </div>
          
          <div className="bg-white rounded-lg shadow p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-500">Last Updated</p>
                <p className="text-2xl font-bold text-gray-900">
                  {summary?.lastUpdated ? new Date(summary.lastUpdated).toLocaleTimeString() : 'N/A'}
                </p>
              </div>
              <LayoutDashboard className="h-8 w-8 text-purple-500" />
            </div>
          </div>
        </div>

        {/* View Mode Tabs */}
        <div className="flex space-x-1 bg-gray-100 p-1 rounded-lg mb-6 w-fit">
          <button
            onClick={() => handleViewModeChange('overview')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'overview' 
                ? 'bg-white text-gray-900 shadow' 
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Overview
          </button>
          <button
            onClick={() => handleViewModeChange('regional')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'regional' 
                ? 'bg-white text-gray-900 shadow' 
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            Regional
          </button>
          <button
            onClick={() => handleViewModeChange('district')}
            className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
              viewMode === 'district' 
                ? 'bg-white text-gray-900 shadow' 
                : 'text-gray-600 hover:text-gray-900'
            }`}
          >
            District
          </button>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content Area */}
          <div className="lg:col-span-2 space-y-6">
            {/* Metric Cards */}
            <section>
              <h2 className="text-lg font-semibold text-gray-900 mb-4">Key Economic Indicators</h2>
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
                {metrics?.map((metric) => (
                  <MetricCard key={metric.id} metric={metric} />
                ))}
              </div>
            </section>

            {/* Regional Heatmap */}
            <section>
              <div className="flex justify-between items-center mb-4">
                <h2 className="text-lg font-semibold text-gray-900">Regional Economic Stress Map</h2>
                <select className="border border-gray-300 rounded-md px-3 py-1 text-sm">
                  <option value="stress">Stress Index</option>
                  <option value="inflation">Inflation</option>
                  <option value="unemployment">Unemployment</option>
                </select>
              </div>
              <RegionHeatmap regions={regions || []} />
            </section>
          </div>

          {/* Side Panel */}
          <div className="space-y-6">
            {/* Quick Actions */}
            <div className="bg-white rounded-lg shadow p-4">
              <h3 className="font-semibold text-gray-900 mb-3">Quick Actions</h3>
              <div className="space-y-2">
                <button 
                  onClick={() => setActivePanel('alerts')}
                  className="w-full flex items-center justify-between p-3 bg-orange-50 rounded-lg hover:bg-orange-100"
                >
                  <div className="flex items-center space-x-3">
                    <Bell className="h-5 w-5 text-orange-600" />
                    <span className="text-sm font-medium text-gray-900">View Active Alerts</span>
                  </div>
                  {summary && summary.activeAlerts > 0 && (
                    <span className="bg-orange-600 text-white text-xs px-2 py-1 rounded-full">
                      {summary.activeAlerts}
                    </span>
                  )}
                </button>
                
                <button 
                  onClick={() => setActivePanel('interventions')}
                  className="w-full flex items-center justify-between p-3 bg-blue-50 rounded-lg hover:bg-blue-100"
                >
                  <div className="flex items-center space-x-3">
                    <TrendingUp className="h-5 w-5 text-blue-600" />
                    <span className="text-sm font-medium text-gray-900">Active Interventions</span>
                  </div>
                  {summary && summary.activeInterventions > 0 && (
                    <span className="bg-blue-600 text-white text-xs px-2 py-1 rounded-full">
                      {summary.activeInterventions}
                    </span>
                  )}
                </button>
                
                <button 
                  onClick={() => navigate('/simulation')}
                  className="w-full flex items-center justify-between p-3 bg-green-50 rounded-lg hover:bg-green-100"
                >
                  <div className="flex items-center space-x-3">
                    <Activity className="h-5 w-5 text-green-600" />
                    <span className="text-sm font-medium text-gray-900">Run Simulation</span>
                  </div>
                </button>
              </div>
            </div>

            {/* Alerts Panel */}
            {(activePanel === 'alerts' || !activePanel) && (
              <AlertsPanel onClose={() => setActivePanel(null)} />
            )}

            {/* Interventions Panel */}
            {activePanel === 'interventions' && (
              <InterventionsPanel onClose={() => setActivePanel(null)} />
            )}
          </div>
        </div>
      </main>

      {/* Export Modal */}
      {showExport && (
        <ExportModal onClose={() => setShowExport(false)} />
      )}
    </div>
  );
};
