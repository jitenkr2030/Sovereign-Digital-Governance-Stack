import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  AlertTriangle,
  Calendar,
  User,
  Shield,
  Tag,
  Save,
  Send,
  X,
} from 'lucide-react';
import { useIncidentStore } from '../store/incidentStore';
import { IncidentSeverity, IncidentType } from '../types/incident';

const CreateIncident: React.FC = () => {
  const navigate = useNavigate();
  const { createIncident } = useIncidentStore();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const [formData, setFormData] = useState({
    title: '',
    description: '',
    severity: 'medium' as IncidentSeverity,
    type: 'technical' as IncidentType,
    affectedSystems: [''],
    tags: [''],
    assigneeId: '',
    dueDate: '',
  });

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({ ...prev, [name]: value }));
  };

  const handleArrayChange = (
    index: number,
    field: 'affectedSystems' | 'tags',
    value: string
  ) => {
    const updated = [...formData[field]];
    updated[index] = value;
    setFormData((prev) => ({ ...prev, [field]: updated }));
  };

  const addArrayItem = (field: 'affectedSystems' | 'tags') => {
    setFormData((prev) => ({ ...prev, [field]: [...prev[field], ''] }));
  };

  const removeArrayItem = (index: number, field: 'affectedSystems' | 'tags') => {
    setFormData((prev) => ({
      ...prev,
      [field]: prev[field].filter((_, i) => i !== index),
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    try {
      const incident = await createIncident({
        title: formData.title,
        description: formData.description,
        severity: formData.severity,
        type: formData.type,
        affectedSystems: formData.affectedSystems.filter((s) => s.trim()),
        tags: formData.tags.filter((t) => t.trim()),
        priority: formData.severity === 'critical' ? 1 : formData.severity === 'high' ? 2 : 3,
        reporter: { id: '1', name: 'Current User', email: 'user@csic.io', role: 'Analyst', department: 'Security' },
      });

      navigate(`/incidents/${incident.id}`);
    } catch (error) {
      console.error('Failed to create incident:', error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const severityOptions: IncidentSeverity[] = ['critical', 'high', 'medium', 'low', 'info'];
  const typeOptions: IncidentType[] = ['security', 'compliance', 'technical', 'operational', 'fraud'];

  return (
    <div className="max-w-4xl mx-auto">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/incidents')}
          className="p-2 rounded-lg hover:bg-gray-800 text-gray-400 hover:text-white transition-colors"
        >
          <ArrowLeft size={20} />
        </button>
        <div>
          <h1 className="text-2xl font-bold text-white">Create New Incident</h1>
          <p className="text-gray-400 mt-1">Report a new security incident or issue</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Information */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h2 className="text-lg font-semibold text-white mb-4">Basic Information</h2>
          
          {/* Title */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-400 mb-2">
              Incident Title *
            </label>
            <input
              type="text"
              name="title"
              value={formData.title}
              onChange={handleInputChange}
              placeholder="Brief description of the incident"
              className="input"
              required
            />
          </div>

          {/* Description */}
          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-400 mb-2">
              Description *
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleInputChange}
              placeholder="Detailed description of what happened, when, and where..."
              className="input min-h-[150px] resize-y"
              required
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Severity */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Severity *
              </label>
              <select
                name="severity"
                value={formData.severity}
                onChange={handleInputChange}
                className="select"
                required
              >
                {severityOptions.map((severity) => (
                  <option key={severity} value={severity}>
                    {severity.charAt(0).toUpperCase() + severity.slice(1)}
                  </option>
                ))}
              </select>
            </div>

            {/* Type */}
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Incident Type *
              </label>
              <select
                name="type"
                value={formData.type}
                onChange={handleInputChange}
                className="select"
                required
              >
                {typeOptions.map((type) => (
                  <option key={type} value={type}>
                    {type.charAt(0).toUpperCase() + type.slice(1)}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* Affected Systems */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h2 className="text-lg font-semibold text-white mb-4">Affected Systems</h2>
          <div className="space-y-3">
            {formData.affectedSystems.map((system, index) => (
              <div key={index} className="flex gap-2">
                <input
                  type="text"
                  value={system}
                  onChange={(e) => handleArrayChange(index, 'affectedSystems', e.target.value)}
                  placeholder="System name or component"
                  className="input flex-1"
                />
                {formData.affectedSystems.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeArrayItem(index, 'affectedSystems')}
                    className="p-2 text-gray-400 hover:text-red-400 transition-colors"
                  >
                    <X size={20} />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('affectedSystems')}
              className="text-sm text-primary-400 hover:text-primary-300"
            >
              + Add another system
            </button>
          </div>
        </div>

        {/* Tags */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h2 className="text-lg font-semibold text-white mb-4">Tags</h2>
          <div className="space-y-3">
            {formData.tags.map((tag, index) => (
              <div key={index} className="flex gap-2">
                <input
                  type="text"
                  value={tag}
                  onChange={(e) => handleArrayChange(index, 'tags', e.target.value)}
                  placeholder="Tag name"
                  className="input flex-1"
                />
                {formData.tags.length > 1 && (
                  <button
                    type="button"
                    onClick={() => removeArrayItem(index, 'tags')}
                    className="p-2 text-gray-400 hover:text-red-400 transition-colors"
                  >
                    <X size={20} />
                  </button>
                )}
              </div>
            ))}
            <button
              type="button"
              onClick={() => addArrayItem('tags')}
              className="text-sm text-primary-400 hover:text-primary-300"
            >
              + Add another tag
            </button>
          </div>
        </div>

        {/* Assignment */}
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h2 className="text-lg font-semibold text-white mb-4">Assignment</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Assign To
              </label>
              <select
                name="assigneeId"
                value={formData.assigneeId}
                onChange={handleInputChange}
                className="select"
              >
                <option value="">Unassigned</option>
                <option value="1">John Doe - Security Analyst</option>
                <option value="2">Jane Smith - Lead Investigator</option>
                <option value="3">Mike Johnson - Compliance Officer</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-400 mb-2">
                Due Date
              </label>
              <input
                type="date"
                name="dueDate"
                value={formData.dueDate}
                onChange={handleInputChange}
                className="input"
              />
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-end gap-4">
          <button
            type="button"
            onClick={() => navigate('/incidents')}
            className="btn btn-secondary"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={isSubmitting}
            className="btn btn-primary flex items-center gap-2"
          >
            {isSubmitting ? (
              <>
                <div className="spinner w-4 h-4" />
                Creating...
              </>
            ) : (
              <>
                <Send size={18} />
                Create Incident
              </>
            )}
          </button>
        </div>
      </form>
    </div>
  );
};

export default CreateIncident;
