import React, { useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import {
  Search,
  Filter,
  Plus,
  SortAsc,
  SortDesc,
  ChevronLeft,
  ChevronRight,
  AlertTriangle,
  Calendar,
  Clock,
  User,
  Tag,
} from 'lucide-react';
import { useIncidentStore } from '../store/incidentStore';
import { Incident, IncidentSeverity, IncidentStatus, IncidentType } from '../types/incident';

const IncidentList: React.FC = () => {
  const { incidents, isLoading, fetchIncidents, setFilters, filters } = useIncidentStore();
  const [searchParams, setSearchParams] = useSearchParams();
  const [searchTerm, setSearchTerm] = useState('');
  const [sortField, setSortField] = useState<'createdAt' | 'severity' | 'status'>('createdAt');
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc');
  const [currentPage, setCurrentPage] = useState(1);
  const [showFilters, setShowFilters] = useState(false);
  const itemsPerPage = 10;

  useEffect(() => {
    fetchIncidents();
  }, [fetchIncidents]);

  const severityOptions: IncidentSeverity[] = ['critical', 'high', 'medium', 'low', 'info'];
  const statusOptions: IncidentStatus[] = ['open', 'investigating', 'containment', 'resolved', 'closed'];
  const typeOptions: IncidentType[] = ['security', 'compliance', 'technical', 'operational', 'fraud'];

  // Filter and search
  const filteredIncidents = incidents.filter((incident) => {
    const matchesSearch =
      !searchTerm ||
      incident.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      incident.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      incident.description.toLowerCase().includes(searchTerm.toLowerCase());

    const matchesSeverity = !filters.severity || filters.severity.includes(incident.severity);
    const matchesStatus = !filters.status || filters.status.includes(incident.status);
    const matchesType = !filters.type || filters.type === incident.type;

    return matchesSearch && matchesSeverity && matchesStatus && matchesType;
  });

  // Sort
  const sortedIncidents = [...filteredIncidents].sort((a, b) => {
    let comparison = 0;
    switch (sortField) {
      case 'severity':
        comparison = getSeverityWeight(a.severity) - getSeverityWeight(b.severity);
        break;
      case 'status':
        comparison = getStatusWeight(a.status) - getStatusWeight(b.status);
        break;
      case 'createdAt':
      default:
        comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime();
    }
    return sortDirection === 'asc' ? comparison : -comparison;
  });

  // Paginate
  const totalPages = Math.ceil(sortedIncidents.length / itemsPerPage);
  const paginatedIncidents = sortedIncidents.slice(
    (currentPage - 1) * itemsPerPage,
    currentPage * itemsPerPage
  );

  const getSeverityWeight = (severity: IncidentSeverity): number => {
    const weights: Record<IncidentSeverity, number> = {
      critical: 5,
      high: 4,
      medium: 3,
      low: 2,
      info: 1,
    };
    return weights[severity];
  };

  const getStatusWeight = (status: IncidentStatus): number => {
    const weights: Record<IncidentStatus, number> = {
      open: 5,
      investigating: 4,
      containment: 3,
      resolved: 2,
      closed: 1,
    };
    return weights[status];
  };

  const handleSort = (field: typeof sortField) => {
    if (sortField === field) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
    } else {
      setSortField(field);
      setSortDirection('desc');
    }
  };

  const severityColors: Record<IncidentSeverity, string> = {
    critical: 'border-l-red-500 bg-red-500/10',
    high: 'border-l-orange-500 bg-orange-500/10',
    medium: 'border-l-yellow-500 bg-yellow-500/10',
    low: 'border-l-green-500 bg-green-500/10',
    info: 'border-l-blue-500 bg-blue-500/10',
  };

  const statusColors: Record<IncidentStatus, string> = {
    open: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    investigating: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
    containment: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    resolved: 'bg-green-500/20 text-green-400 border-green-500/30',
    closed: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold text-white">Incidents</h1>
          <p className="text-gray-400 mt-1">
            {filteredIncidents.length} incident{filteredIncidents.length !== 1 ? 's' : ''} found
          </p>
        </div>
        <Link
          to="/incidents/new"
          className="btn btn-primary flex items-center gap-2"
        >
          <Plus size={18} />
          New Incident
        </Link>
      </div>

      {/* Search and Filters */}
      <div className="bg-gray-800 rounded-xl p-4 border border-gray-700">
        <div className="flex flex-col lg:flex-row gap-4">
          {/* Search */}
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
            <input
              type="text"
              placeholder="Search incidents by title, ID, or description..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="input pl-10"
            />
          </div>

          {/* Filter Toggle */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`btn btn-secondary flex items-center gap-2 ${showFilters ? 'bg-gray-600' : ''}`}
          >
            <Filter size={18} />
            Filters
            {(filters.severity?.length || filters.status?.length || filters.type) && (
              <span className="w-2 h-2 bg-primary-500 rounded-full" />
            )}
          </button>
        </div>

        {/* Filter Panel */}
        {showFilters && (
          <div className="mt-4 pt-4 border-t border-gray-700 grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Severity Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Severity</label>
              <div className="flex flex-wrap gap-2">
                {severityOptions.map((severity) => (
                  <button
                    key={severity}
                    onClick={() => {
                      const current = filters.severity || [];
                      const updated = current.includes(severity)
                        ? current.filter((s) => s !== severity)
                        : [...current, severity];
                      setFilters({ ...filters, severity: updated });
                    }}
                    className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                      filters.severity?.includes(severity)
                        ? 'bg-gray-600 border-gray-500 text-white'
                        : 'bg-gray-900 border-gray-700 text-gray-400 hover:border-gray-500'
                    }`}
                  >
                    {severity}
                  </button>
                ))}
              </div>
            </div>

            {/* Status Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Status</label>
              <div className="flex flex-wrap gap-2">
                {statusOptions.map((status) => (
                  <button
                    key={status}
                    onClick={() => {
                      const current = filters.status || [];
                      const updated = current.includes(status)
                        ? current.filter((s) => s !== status)
                        : [...current, status];
                      setFilters({ ...filters, status: updated });
                    }}
                    className={`px-3 py-1 rounded-full text-xs font-medium border transition-colors ${
                      filters.status?.includes(status)
                        ? 'bg-gray-600 border-gray-500 text-white'
                        : 'bg-gray-900 border-gray-700 text-gray-400 hover:border-gray-500'
                    }`}
                  >
                    {status}
                  </button>
                ))}
              </div>
            </div>

            {/* Type Filter */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">Type</label>
              <select
                value={filters.type || ''}
                onChange={(e) => setFilters({ ...filters, type: e.target.value as IncidentType || undefined })}
                className="select"
              >
                <option value="">All Types</option>
                {typeOptions.map((type) => (
                  <option key={type} value={type}>
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </option>
                ))}
              </select>
            </div>
          </div>
        )}
      </div>

      {/* Incidents Table */}
      <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
        {/* Table Header */}
        <div className="grid grid-cols-12 gap-4 px-4 py-3 bg-gray-900/50 border-b border-gray-700 text-sm font-medium text-gray-400">
          <div className="col-span-4">Incident</div>
          <div className="col-span-2">Severity</div>
          <div className="col-span-2">Status</div>
          <div className="col-span-2">
            <button
              onClick={() => handleSort('createdAt')}
              className="flex items-center gap-1 hover:text-white"
            >
              Created
              {sortField === 'createdAt' &&
                (sortDirection === 'asc' ? <SortAsc size={14} /> : <SortDesc size={14} />)}
            </button>
          </div>
          <div className="col-span-2 text-right">Actions</div>
        </div>

        {/* Table Body */}
        {isLoading ? (
          <div className="p-8 text-center">
            <div className="spinner mx-auto mb-4" />
            <p className="text-gray-400">Loading incidents...</p>
          </div>
        ) : paginatedIncidents.length === 0 ? (
          <div className="empty-state py-12">
            <AlertTriangle className="empty-state-icon" />
            <h3 className="empty-state-title">No incidents found</h3>
            <p className="empty-state-description">
              {searchTerm || filters.severity?.length || filters.status?.length || filters.type
                ? 'Try adjusting your search or filters'
                : 'Get started by creating your first incident'}
            </p>
            <Link to="/incidents/new" className="btn btn-primary mt-4">
              Create Incident
            </Link>
          </div>
        ) : (
          <div className="divide-y divide-gray-700">
            {paginatedIncidents.map((incident) => (
              <Link
                key={incident.id}
                to={`/incidents/${incident.id}`}
                className={`grid grid-cols-12 gap-4 px-4 py-4 hover:bg-gray-700/30 transition-colors border-l-4 ${severityColors[incident.severity]}`}
              >
                <div className="col-span-4">
                  <div className="flex items-start gap-3">
                    <span className="font-mono text-sm text-primary-400">{incident.id}</span>
                    <div>
                      <h3 className="font-medium text-white">{incident.title}</h3>
                      <p className="text-sm text-gray-400 truncate mt-0.5">{incident.description}</p>
                      <div className="flex items-center gap-2 mt-2">
                        <Tag size={12} className="text-gray-500" />
                        {incident.tags.slice(0, 3).map((tag) => (
                          <span
                            key={tag}
                            className="px-2 py-0.5 bg-gray-700 rounded text-xs text-gray-300"
                          >
                            {tag}
                          </span>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
                <div className="col-span-2 flex items-center">
                  <span className={`status-badge status-${incident.severity}`}>
                    {incident.severity}
                  </span>
                </div>
                <div className="col-span-2 flex items-center">
                  <span
                    className={`status-badge border ${statusColors[incident.status]}`}
                  >
                    {incident.status}
                  </span>
                </div>
                <div className="col-span-2 flex items-center text-sm text-gray-400">
                  <Clock size={14} className="mr-1.5" />
                  {new Date(incident.createdAt).toLocaleDateString()}
                </div>
                <div className="col-span-2 flex items-center justify-end gap-3">
                  <span className="text-sm text-gray-500 capitalize">{incident.type}</span>
                  <span className="text-primary-400 text-sm font-medium">View →</span>
                </div>
              </Link>
            ))}
          </div>
        )}

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="px-4 py-3 bg-gray-900/50 border-t border-gray-700 flex items-center justify-between">
            <p className="text-sm text-gray-400">
              Showing {(currentPage - 1) * itemsPerPage + 1} to{' '}
              {Math.min(currentPage * itemsPerPage, filteredIncidents.length)} of{' '}
              {filteredIncidents.length} incidents
            </p>
            <div className="flex items-center gap-2">
              <button
                onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
                disabled={currentPage === 1}
                className="p-2 rounded-lg hover:bg-gray-700 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={18} />
              </button>
              {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
                const page = i + 1;
                return (
                  <button
                    key={page}
                    onClick={() => setCurrentPage(page)}
                    className={`w-8 h-8 rounded-lg text-sm font-medium transition-colors ${
                      currentPage === page
                        ? 'bg-primary-600 text-white'
                        : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                    }`}
                  >
                    {page}
                  </button>
                );
              })}
              <button
                onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages}
                className="p-2 rounded-lg hover:bg-gray-700 text-gray-400 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <ChevronRight size={18} />
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default IncidentList;
