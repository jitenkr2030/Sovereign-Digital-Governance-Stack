import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
  ArrowLeft,
  Edit,
  User,
  Clock,
  Tag,
  AlertTriangle,
  Shield,
  FileText,
  MessageSquare,
  CheckCircle,
  Send,
  MoreHorizontal,
  ExternalLink,
} from 'lucide-react';
import { useIncidentStore } from '../store/incidentStore';
import { Incident, TimelineEvent } from '../types/incident';

const IncidentDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const { selectedIncident, fetchIncident, isLoading, updateIncident } = useIncidentStore();
  const [activeTab, setActiveTab] = useState<'overview' | 'timeline' | 'ioc' | 'related'>('overview');
  const [newComment, setNewComment] = useState('');

  useEffect(() => {
    if (id) {
      fetchIncident(id);
    }
  }, [id, fetchIncident]);

  if (isLoading || !selectedIncident) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="spinner" />
      </div>
    );
  }

  const incident = selectedIncident;

  const severityColors: Record<string, string> = {
    critical: 'bg-red-500/20 text-red-400 border-red-500/30',
    high: 'bg-orange-500/20 text-orange-400 border-orange-500/30',
    medium: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    low: 'bg-green-500/20 text-green-400 border-green-500/30',
    info: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
  };

  const statusColors: Record<string, string> = {
    open: 'bg-blue-500/20 text-blue-400 border-blue-500/30',
    investigating: 'bg-purple-500/20 text-purple-400 border-purple-500/30',
    containment: 'bg-yellow-500/20 text-yellow-400 border-yellow-500/30',
    resolved: 'bg-green-500/20 text-green-400 border-green-500/30',
    closed: 'bg-gray-500/20 text-gray-400 border-gray-500/30',
  };

  const handleStatusChange = async (newStatus: string) => {
    if (id) {
      await updateIncident(id, { status: newStatus as Incident['status'] });
    }
  };

  const handleAddComment = () => {
    if (newComment.trim()) {
      // Implementation would add comment to timeline
      setNewComment('');
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-start gap-4">
          <Link
            to="/incidents"
            className="p-2 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white transition-colors"
          >
            <ArrowLeft size={20} />
          </Link>
          <div>
            <div className="flex items-center gap-3">
              <span className="font-mono text-sm text-primary-400">{incident.id}</span>
              <span className={`status-badge border ${severityColors[incident.severity]}`}>
                {incident.severity}
              </span>
              <span className={`status-badge border ${statusColors[incident.status]}`}>
                {incident.status}
              </span>
            </div>
            <h1 className="text-2xl font-bold text-white mt-2">{incident.title}</h1>
            <div className="flex items-center gap-4 mt-2 text-sm text-gray-400">
              <span className="flex items-center gap-1">
                <Clock size={14} />
                Created {new Date(incident.createdAt).toLocaleString()}
              </span>
              <span className="flex items-center gap-1">
                <User size={14} />
                {incident.reporter.name}
              </span>
              <span className="capitalize">{incident.type}</span>
            </div>
          </div>
        </div>
        <div className="flex items-center gap-3">
          <select
            value={incident.status}
            onChange={(e) => handleStatusChange(e.target.value)}
            className="select w-40"
          >
            <option value="open">Open</option>
            <option value="investigating">Investigating</option>
            <option value="containment">Containment</option>
            <option value="resolved">Resolved</option>
            <option value="closed">Closed</option>
          </select>
          <button className="btn btn-secondary flex items-center gap-2">
            <Edit size={16} />
            Edit
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-700">
        <nav className="flex gap-6">
          {[
            { id: 'overview', label: 'Overview', icon: FileText },
            { id: 'timeline', label: 'Timeline', icon: Clock },
            { id: 'ioc', label: 'Indicators', icon: AlertTriangle },
            { id: 'related', label: 'Related', icon: Link },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as typeof activeTab)}
              className={`flex items-center gap-2 pb-3 px-1 border-b-2 transition-colors ${
                activeTab === tab.id
                  ? 'border-primary-500 text-primary-400'
                  : 'border-transparent text-gray-400 hover:text-white'
              }`}
            >
              <tab.icon size={16} />
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Tab Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Content */}
        <div className="lg:col-span-2 space-y-6">
          {activeTab === 'overview' && (
            <>
              {/* Description */}
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4">Description</h3>
                <p className="text-gray-300 leading-relaxed">{incident.description}</p>
              </div>

              {/* Affected Systems */}
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4">Affected Systems</h3>
                <div className="flex flex-wrap gap-2">
                  {incident.affectedSystems.map((system) => (
                    <span
                      key={system}
                      className="px-3 py-1.5 bg-gray-700 rounded-lg text-sm text-gray-300 flex items-center gap-2"
                    >
                      <Shield size={14} className="text-primary-400" />
                      {system}
                    </span>
                  ))}
                </div>
              </div>

              {/* Metrics */}
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4">Response Metrics</h3>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                  <div className="text-center p-4 bg-gray-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">{incident.metrics.timeToDetect}m</p>
                    <p className="text-sm text-gray-400">Time to Detect</p>
                  </div>
                  <div className="text-center p-4 bg-gray-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">{incident.metrics.timeToRespond}m</p>
                    <p className="text-sm text-gray-400">Time to Respond</p>
                  </div>
                  <div className="text-center p-4 bg-gray-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">
                      {incident.metrics.timeToResolve > 0
                        ? `${Math.floor(incident.metrics.timeToResolve / 60)}h ${incident.metrics.timeToResolve % 60}m`
                        : 'N/A'}
                    </p>
                    <p className="text-sm text-gray-400">Time to Resolve</p>
                  </div>
                  <div className="text-center p-4 bg-gray-700/50 rounded-lg">
                    <p className="text-2xl font-bold text-white">${incident.metrics.estimatedImpact.toLocaleString()}</p>
                    <p className="text-sm text-gray-400">Est. Impact</p>
                  </div>
                </div>
              </div>

              {/* Tags */}
              <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
                <h3 className="text-lg font-semibold text-white mb-4">Tags</h3>
                <div className="flex flex-wrap gap-2">
                  {incident.tags.map((tag) => (
                    <span
                      key={tag}
                      className="px-3 py-1 bg-primary-500/20 text-primary-400 rounded-full text-sm"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              </div>
            </>
          )}

          {activeTab === 'timeline' && (
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h3 className="text-lg font-semibold text-white mb-6">Incident Timeline</h3>

              {/* Add Comment */}
              <div className="mb-6 pb-6 border-b border-gray-700">
                <div className="flex gap-3">
                  <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center flex-shrink-0">
                    <User size={20} className="text-gray-400" />
                  </div>
                  <div className="flex-1">
                    <textarea
                      value={newComment}
                      onChange={(e) => setNewComment(e.target.value)}
                      placeholder="Add a comment or update..."
                      className="input min-h-[80px] resize-none"
                    />
                    <div className="flex justify-end mt-2">
                      <button
                        onClick={handleAddComment}
                        disabled={!newComment.trim()}
                        className="btn btn-primary flex items-center gap-2"
                      >
                        <Send size={16} />
                        Add Update
                      </button>
                    </div>
                  </div>
                </div>
              </div>

              {/* Timeline */}
              <div className="space-y-0">
                {incident.timeline.length > 0 ? (
                  incident.timeline.map((event) => (
                    <div key={event.id} className="timeline-item">
                      <div className="timeline-dot">
                        <Clock size={12} className="text-gray-400" />
                      </div>
                      <div>
                        <div className="flex items-center gap-2 mb-1">
                          <span className="font-medium text-white">{event.actor.name}</span>
                          <span className="text-xs px-2 py-0.5 bg-gray-700 rounded text-gray-400 capitalize">
                            {event.type}
                          </span>
                        </div>
                        <p className="text-gray-300">{event.description}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {new Date(event.timestamp).toLocaleString()}
                        </p>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="text-center py-8 text-gray-500">
                    <Clock className="w-12 h-12 mx-auto mb-3 opacity-50" />
                    <p>No timeline events yet</p>
                  </div>
                )}

                {/* Creation Event */}
                <div className="timeline-item">
                  <div className="timeline-dot bg-primary-500 border-primary-500">
                    <FileText size={12} className="text-white" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-white">{incident.reporter.name}</span>
                      <span className="text-xs px-2 py-0.5 bg-gray-700 rounded text-gray-400">
                        creation
                      </span>
                    </div>
                    <p className="text-gray-300">Incident reported</p>
                    <p className="text-xs text-gray-500 mt-1">
                      {new Date(incident.createdAt).toLocaleString()}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'ioc' && (
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <div className="flex items-center justify-between mb-6">
                <h3 className="text-lg font-semibold text-white">Indicators of Compromise</h3>
                <button className="btn btn-secondary text-sm">Add IOC</button>
              </div>

              {incident.indicators.length > 0 ? (
                <div className="space-y-3">
                  {incident.indicators.map((ioc) => (
                    <div
                      key={ioc.id}
                      className="p-4 bg-gray-700/50 rounded-lg border border-gray-600"
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className="px-2 py-0.5 bg-primary-500/20 text-primary-400 rounded text-xs uppercase">
                          {ioc.type}
                        </span>
                        <span className="text-sm text-gray-400">
                          Confidence: {ioc.confidence}%
                        </span>
                      </div>
                      <code className="text-sm text-white font-mono block bg-gray-800 p-2 rounded mt-2">
                        {ioc.value}
                      </code>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <AlertTriangle className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p>No indicators of compromise recorded</p>
                </div>
              )}
            </div>
          )}

          {activeTab === 'related' && (
            <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
              <h3 className="text-lg font-semibold text-white mb-6">Related Incidents</h3>

              {incident.relatedIncidents.length > 0 ? (
                <div className="space-y-3">
                  {incident.relatedIncidents.map((relId) => (
                    <Link
                      key={relId}
                      to={`/incidents/${relId}`}
                      className="flex items-center justify-between p-4 bg-gray-700/50 rounded-lg hover:bg-gray-700 transition-colors"
                    >
                      <div className="flex items-center gap-3">
                        <span className="font-mono text-sm text-primary-400">{relId}</span>
                        <span className="text-gray-400">Related incident</span>
                      </div>
                      <ExternalLink size={16} className="text-gray-500" />
                    </Link>
                  ))}
                </div>
              ) : (
                <div className="text-center py-12 text-gray-500">
                  <Link className="w-12 h-12 mx-auto mb-3 opacity-50" />
                  <p>No related incidents linked</p>
                  <button className="btn btn-secondary mt-4">Link Incident</button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Quick Actions */}
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <h3 className="text-lg font-semibold text-white mb-4">Quick Actions</h3>
            <div className="space-y-2">
              <button className="w-full btn btn-secondary justify-start gap-3">
                <User size={18} />
                Assign Team Member
              </button>
              <button className="w-full btn btn-secondary justify-start gap-3">
                <Shield size={18} />
                Escalate
              </button>
              <button className="w-full btn btn-secondary justify-start gap-3">
                <FileText size={18} />
                Generate Report
              </button>
              <button className="w-full btn btn-danger justify-start gap-3">
                <CheckCircle size={18} />
                Mark as Resolved
              </button>
            </div>
          </div>

          {/* Assignee */}
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <h3 className="text-lg font-semibold text-white mb-4">Assigned To</h3>
            {incident.assignee ? (
              <div className="flex items-center gap-3">
                <div className="w-12 h-12 rounded-full bg-gray-700 flex items-center justify-center">
                  <User size={24} className="text-gray-400" />
                </div>
                <div>
                  <p className="font-medium text-white">{incident.assignee.name}</p>
                  <p className="text-sm text-gray-400">{incident.assignee.role}</p>
                </div>
              </div>
            ) : (
              <button className="w-full p-4 border-2 border-dashed border-gray-600 rounded-lg text-gray-400 hover:border-gray-500 hover:text-gray-300 transition-colors">
                + Assign team member
              </button>
            )}
          </div>

          {/* Reporter */}
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <h3 className="text-lg font-semibold text-white mb-4">Reported By</h3>
            <div className="flex items-center gap-3">
              <div className="w-12 h-12 rounded-full bg-gray-700 flex items-center justify-center">
                <User size={24} className="text-gray-400" />
              </div>
              <div>
                <p className="font-medium text-white">{incident.reporter.name}</p>
                <p className="text-sm text-gray-400">{incident.reporter.department}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default IncidentDetail;
