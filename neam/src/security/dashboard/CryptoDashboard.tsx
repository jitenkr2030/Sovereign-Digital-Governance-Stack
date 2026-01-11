/**
 * Cryptographic Operations Dashboard
 * 
 * This module provides a comprehensive dashboard for monitoring and managing
 * cryptographic operations including HSM status, key inventory, certificate
 * management, and audit logs.
 * 
 * @packageDocumentation
 */

import React, { useState, useEffect, useCallback } from 'react';
import { 
  Shield, 
  Key, 
  Certificate, 
  Activity, 
  AlertTriangle, 
  CheckCircle, 
  XCircle, 
  RefreshCw, 
  Plus, 
  Download, 
  Upload,
  Settings,
  Lock,
  Unlock,
  Eye,
  Trash2,
  RotateCw,
  Search,
  Filter,
  DownloadCloud,
  Clock,
  TrendingUp,
  Server,
  Database,
  Fingerprint
} from 'lucide-react';

// Types for dashboard data
interface HSMStatus {
  connected: boolean;
  slotId: number;
  manufacturer: string;
  model: string;
  firmware: string;
  serialNumber: string;
  totalKeys: number;
  certificates: number;
  lastHeartbeat: Date;
  temperature?: number;
  load?: number;
}

interface KeyInventoryItem {
  id: string;
  label: string;
  type: string;
  algorithm: string;
  size: number;
  status: 'active' | 'rotating' | 'retired' | 'compromised';
  createdAt: Date;
  expiresAt?: Date;
  usage: string[];
  lastUsed?: Date;
}

interface CertificateItem {
  id: string;
  commonName: string;
  issuer: string;
  type: string;
  status: 'valid' | 'expired' | 'revoked' | 'pending';
  serialNumber: string;
  fingerprint: string;
  notBefore: Date;
  notAfter: Date;
  daysUntilExpiry: number;
  sans?: string[];
}

interface AuditLogEntry {
  id: string;
  timestamp: Date;
  operation: string;
  severity: 'info' | 'warning' | 'error' | 'critical';
  keyId?: string;
  keyLabel?: string;
  message: string;
  duration: number;
}

interface CryptoStats {
  totalKeys: number;
  activeKeys: number;
  expiringCertificates: number;
  totalSignatures: number;
  signaturesLast24h: number;
  encryptionsLast24h: number;
  averageLatency: number;
}

// Dashboard Component
export const CryptoDashboard: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'keys' | 'certificates' | 'audit'>('overview');
  const [isLoading, setIsLoading] = useState(true);
  const [hsmStatus, setHsmStatus] = useState<HSMStatus | null>(null);
  const [keyInventory, setKeyInventory] = useState<KeyInventoryItem[]>([]);
  const [certificates, setCertificates] = useState<CertificateItem[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [stats, setStats] = useState<CryptoStats | null>(null);
  const [error, setError] = useState<string | null>(null);

  // Fetch dashboard data
  const fetchDashboardData = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      // Simulated API calls - replace with actual endpoints
      const [statusRes, keysRes, certsRes, logsRes, statsRes] = await Promise.all([
        fetch('/api/security/hsm/status').catch(() => null),
        fetch('/api/security/keys').catch(() => null),
        fetch('/api/security/certificates').catch(() => null),
        fetch('/api/security/audit').catch(() => null),
        fetch('/api/security/stats').catch(() => null)
      ]);

      // Mock data for demonstration
      setHsmStatus({
        connected: true,
        slotId: 0,
        manufacturer: 'SoftHSM',
        model: 'v2',
        firmware: '2.6.0',
        serialNumber: 'HSM-001',
        totalKeys: 15,
        certificates: 8,
        lastHeartbeat: new Date(),
        temperature: 45,
        load: 23
      });

      setKeyInventory([
        {
          id: 'key-001',
          label: 'Master Encryption Key',
          type: 'AES',
          algorithm: 'AES-256-GCM',
          size: 256,
          status: 'active',
          createdAt: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
          usage: ['encrypt', 'decrypt', 'wrap', 'unwrap'],
          lastUsed: new Date(Date.now() - 1 * 60 * 60 * 1000)
        },
        {
          id: 'key-002',
          label: 'Document Signing Key',
          type: 'RSA',
          algorithm: 'RSA-SHA256',
          size: 4096,
          status: 'active',
          createdAt: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
          expiresAt: new Date(Date.now() + 300 * 24 * 60 * 60 * 1000),
          usage: ['sign', 'verify'],
          lastUsed: new Date(Date.now() - 2 * 60 * 60 * 1000)
        },
        {
          id: 'key-003',
          label: 'Legacy TLS Key',
          type: 'RSA',
          algorithm: 'RSA-SHA256',
          size: 2048,
          status: 'rotating',
          createdAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000),
          expiresAt: new Date(Date.now() + 270 * 24 * 60 * 60 * 1000),
          usage: ['sign'],
          lastUsed: new Date(Date.now() - 24 * 60 * 60 * 1000)
        }
      ]);

      setCertificates([
        {
          id: 'cert-001',
          commonName: '*.national-dashboard.gov',
          issuer: 'National Dashboard CA',
          type: 'TLS Server',
          status: 'valid',
          serialNumber: '1A2B3C4D5E',
          fingerprint: 'SHA256:ABC123',
          notBefore: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
          notAfter: new Date(Date.now() + 335 * 24 * 60 * 60 * 1000),
          daysUntilExpiry: 335,
          sans: ['*.national-dashboard.gov', 'national-dashboard.gov']
        },
        {
          id: 'cert-002',
          commonName: 'API Client Certificate',
          issuer: 'National Dashboard CA',
          type: 'TLS Client',
          status: 'valid',
          serialNumber: '6F7G8H9I0J',
          fingerprint: 'SHA256:DEF456',
          notBefore: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
          notAfter: new Date(Date.now() + 305 * 24 * 60 * 60 * 1000),
          daysUntilExpiry: 305
        },
        {
          id: 'cert-003',
          commonName: 'Code Signing Certificate',
          issuer: 'National Dashboard CA',
          type: 'Code Signing',
          status: 'expired',
          serialNumber: '1K2L3M4N5O',
          fingerprint: 'SHA256:GHI789',
          notBefore: new Date(Date.now() - 400 * 24 * 60 * 60 * 1000),
          notAfter: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000),
          daysUntilExpiry: -10
        }
      ]);

      setAuditLogs([
        {
          id: 'log-001',
          timestamp: new Date(Date.now() - 1000),
          operation: 'SIGN',
          severity: 'info',
          keyId: 'key-002',
          keyLabel: 'Document Signing Key',
          message: 'Document signed successfully',
          duration: 45
        },
        {
          id: 'log-002',
          timestamp: new Date(Date.now() - 5000),
          operation: 'DECRYPT',
          severity: 'info',
          keyId: 'key-001',
          keyLabel: 'Master Encryption Key',
          message: 'Data decrypted successfully',
          duration: 32
        },
        {
          id: 'log-003',
          timestamp: new Date(Date.now() - 15000),
          operation: 'VERIFY',
          severity: 'warning',
          keyId: 'key-002',
          keyLabel: 'Document Signing Key',
          message: 'Verification attempted with expired certificate',
          duration: 28
        }
      ]);

      setStats({
        totalKeys: 15,
        activeKeys: 12,
        expiringCertificates: 2,
        totalSignatures: 1250,
        signaturesLast24h: 145,
        encryptionsLast24h: 89,
        averageLatency: 42
      });

    } catch (err: any) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchDashboardData();
    // Refresh every 30 seconds
    const interval = setInterval(fetchDashboardData, 30000);
    return () => clearInterval(interval);
  }, [fetchDashboardData]);

  // Render overview tab
  const renderOverview = () => (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
      {/* HSM Status Card */}
      <StatusCard
        title="HSM Status"
        icon={<Server className="w-6 h-6" />}
        status={hsmStatus?.connected ? 'healthy' : 'warning'}
        subtitle={hsmStatus?.connected ? 'Connected' : 'Disconnected'}
        details={[
          { label: 'Manufacturer', value: hsmStatus?.manufacturer || 'N/A' },
          { label: 'Model', value: hsmStatus?.model || 'N/A' },
          { label: 'Temperature', value: hsmStatus?.temperature ? `${hsmStatus.temperature}°C` : 'N/A' },
          { label: 'Load', value: hsmStatus?.load ? `${hsmStatus.load}%` : 'N/A' }
        ]}
      />

      {/* Keys Card */}
      <StatusCard
        title="Key Inventory"
        icon={<Key className="w-6 h-6" />}
        status="healthy"
        subtitle={`${stats?.activeKeys || 0} active keys`}
        details={[
          { label: 'Total Keys', value: stats?.totalKeys || 0 },
          { label: 'Active', value: stats?.activeKeys || 0 },
          { label: 'Rotating', value: keyInventory.filter(k => k.status === 'rotating').length },
          { label: 'Retired', value: keyInventory.filter(k => k.status === 'retired').length }
        ]}
      />

      {/* Certificates Card */}
      <StatusCard
        title="Certificates"
        icon={<Certificate className="w-6 h-6" />}
        status={stats?.expiringCertificates ? 'warning' : 'healthy'}
        subtitle={`${stats?.expiringCertificates || 0} expiring soon`}
        details={[
          { label: 'Valid', value: certificates.filter(c => c.status === 'valid').length },
          { label: 'Expired', value: certificates.filter(c => c.status === 'expired').length },
          { label: 'Revoked', value: certificates.filter(c => c.status === 'revoked').length },
          { label: 'Pending', value: certificates.filter(c => c.status === 'pending').length }
        ]}
      />

      {/* Operations Card */}
      <StatusCard
        title="Operations (24h)"
        icon={<Activity className="w-6 h-6" />}
        status="healthy"
        subtitle={`Avg latency: ${stats?.averageLatency || 0}ms`}
        details={[
          { label: 'Signatures', value: stats?.signaturesLast24h || 0 },
          { label: 'Encryptions', value: stats?.encryptionsLast24h || 0 },
          { label: 'Total Signatures', value: stats?.totalSignatures?.toLocaleString() || 0 },
          { label: 'Avg Latency', value: `${stats?.averageLatency || 0}ms` }
        ]}
      />
    </div>
  );

  // Render keys tab
  const renderKeys = () => (
    <div className="bg-white rounded-lg shadow">
      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900">Key Inventory</h3>
        <div className="flex space-x-3">
          <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            <Plus className="w-4 h-4 mr-2" />
            Generate Key
          </button>
          <button className="flex items-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
            <Download className="w-4 h-4 mr-2" />
            Export
          </button>
        </div>
      </div>
      
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key Label</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Algorithm</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {keyInventory.map((key) => (
              <tr key={key.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    <Lock className="w-4 h-4 text-gray-400 mr-3" />
                    <span className="text-sm font-medium text-gray-900">{key.label}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{key.type}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{key.algorithm}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <StatusBadge status={key.status} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {key.createdAt.toLocaleDateString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  <div className="flex space-x-2">
                    <button className="text-blue-600 hover:text-blue-800" title="View Details">
                      <Eye className="w-4 h-4" />
                    </button>
                    <button className="text-green-600 hover:text-green-800" title="Rotate Key">
                      <RotateCw className="w-4 h-4" />
                    </button>
                    <button className="text-red-600 hover:text-red-800" title="Retire Key">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  // Render certificates tab
  const renderCertificates = () => (
    <div className="bg-white rounded-lg shadow">
      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900">Certificate Management</h3>
        <div className="flex space-x-3">
          <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            <Plus className="w-4 h-4 mr-2" />
            Issue Certificate
          </button>
          <button className="flex items-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
            <DownloadCloud className="w-4 h-4 mr-2" />
            Export All
          </button>
        </div>
      </div>
      
      <div className="overflow-x-auto">
        <table className="w-full">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Common Name</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Issuer</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Expires</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {certificates.map((cert) => (
              <tr key={cert.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    <Certificate className="w-4 h-4 text-gray-400 mr-3" />
                    <div>
                      <span className="text-sm font-medium text-gray-900">{cert.commonName}</span>
                      <span className="text-xs text-gray-500 block">{cert.serialNumber}</span>
                    </div>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{cert.issuer}</td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{cert.type}</td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <StatusBadge status={cert.status} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className={`text-sm ${cert.daysUntilExpiry < 30 ? 'text-red-600 font-medium' : 'text-gray-500'}`}>
                    {cert.daysUntilExpiry < 0 
                      ? `Expired ${Math.abs(cert.daysUntilExpiry)} days ago`
                      : `In ${cert.daysUntilExpiry} days`
                    }
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  <div className="flex space-x-2">
                    <button className="text-blue-600 hover:text-blue-800" title="View Details">
                      <Eye className="w-4 h-4" />
                    </button>
                    <button className="text-green-600 hover:text-green-800" title="Download">
                      <Download className="w-4 h-4" />
                    </button>
                    <button className="text-orange-600 hover:text-orange-800" title="Renew">
                      <RefreshCw className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  // Render audit tab
  const renderAudit = () => (
    <div className="bg-white rounded-lg shadow">
      <div className="px-6 py-4 border-b border-gray-200 flex justify-between items-center">
        <h3 className="text-lg font-semibold text-gray-900">Audit Log</h3>
        <div className="flex space-x-3">
          <div className="relative">
            <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
            <input
              type="text"
              placeholder="Search logs..."
              className="pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
          <select className="px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500">
            <option value="">All Operations</option>
            <option value="SIGN">Sign</option>
            <option value="VERIFY">Verify</option>
            <option value="ENCRYPT">Encrypt</option>
            <option value="DECRYPT">Decrypt</option>
            <option value="GENERATE">Generate</option>
          </select>
          <button className="flex items-center px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50">
            <Download className="w-4 h-4 mr-2" />
            Export
          </button>
        </div>
      </div>
      
      <div className="overflow-x-auto max-h-[600px] overflow-y-auto">
        <table className="w-full">
          <thead className="bg-gray-50 sticky top-0">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Operation</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Severity</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message</th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Duration</th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {auditLogs.map((log) => (
              <tr key={log.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {log.timestamp.toLocaleString()}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                  {log.operation}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <SeverityBadge severity={log.severity} />
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {log.keyLabel || log.keyId}
                </td>
                <td className="px-6 py-4 text-sm text-gray-500 max-w-md truncate">
                  {log.message}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {log.duration}ms
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-center">
          <RefreshCw className="w-12 h-12 text-blue-600 animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading cryptographic operations dashboard...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md">
          <AlertTriangle className="w-12 h-12 text-red-500 mx-auto mb-4" />
          <h2 className="text-xl font-semibold text-gray-900 text-center mb-2">Connection Error</h2>
          <p className="text-gray-600 text-center mb-4">{error}</p>
          <button 
            onClick={fetchDashboardData}
            className="w-full py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Retry Connection
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      {/* Header */}
      <div className="mb-8 flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 flex items-center">
            <Shield className="w-10 h-10 text-blue-600 mr-3" />
            Cryptographic Operations
          </h1>
          <p className="text-gray-600 mt-1">HSM key management and cryptographic operations monitoring</p>
        </div>
        <div className="flex space-x-3">
          <button 
            onClick={fetchDashboardData}
            className="flex items-center px-4 py-2 bg-white border border-gray-300 rounded-lg hover:bg-gray-50"
          >
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </button>
          <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
            <Settings className="w-4 h-4 mr-2" />
            Settings
          </button>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="mb-6">
        <nav className="flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: <Activity className="w-4 h-4 mr-2" /> },
            { id: 'keys', label: 'Key Inventory', icon: <Key className="w-4 h-4 mr-2" /> },
            { id: 'certificates', label: 'Certificates', icon: <Certificate className="w-4 h-4 mr-2" /> },
            { id: 'audit', label: 'Audit Log', icon: <Fingerprint className="w-4 h-4 mr-2" /> }
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center px-4 py-2 border-b-2 font-medium text-sm transition-colors ${
                activeTab === tab.id
                  ? 'border-blue-600 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              {tab.icon}
              {tab.label}
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      {activeTab === 'overview' && renderOverview()}
      {activeTab === 'keys' && renderKeys()}
      {activeTab === 'certificates' && renderCertificates()}
      {activeTab === 'audit' && renderAudit()}
    </div>
  );
};

// Status Card Component
interface StatusCardProps {
  title: string;
  icon: React.ReactNode;
  status: 'healthy' | 'warning' | 'critical';
  subtitle: string;
  details: { label: string; value: string | number }[];
}

const StatusCard: React.FC<StatusCardProps> = ({ title, icon, status, subtitle, details }) => {
  const statusColors = {
    healthy: 'bg-green-50 border-green-200',
    warning: 'bg-yellow-50 border-yellow-200',
    critical: 'bg-red-50 border-red-200'
  };

  const iconColors = {
    healthy: 'text-green-600',
    warning: 'text-yellow-600',
    critical: 'text-red-600'
  };

  return (
    <div className={`rounded-lg border ${statusColors[status]} p-6`}>
      <div className="flex items-center justify-between mb-4">
        <div className={`${iconColors[status]}`}>{icon}</div>
        <div className={`w-3 h-3 rounded-full ${
          status === 'healthy' ? 'bg-green-500' : 
          status === 'warning' ? 'bg-yellow-500' : 'bg-red-500'
        }`} />
      </div>
      <h3 className="text-lg font-semibold text-gray-900 mb-1">{title}</h3>
      <p className="text-sm text-gray-600 mb-4">{subtitle}</p>
      <dl className="space-y-2">
        {details.map((detail, index) => (
          <div key={index} className="flex justify-between text-sm">
            <dt className="text-gray-500">{detail.label}</dt>
            <dd className="font-medium text-gray-900">{detail.value}</dd>
          </div>
        ))}
      </dl>
    </div>
  );
};

// Status Badge Component
interface StatusBadgeProps {
  status: string;
}

const StatusBadge: React.FC<StatusBadgeProps> = ({ status }) => {
  const statusConfig: Record<string, { bg: string; text: string }> = {
    active: { bg: 'bg-green-100', text: 'text-green-800' },
    rotating: { bg: 'bg-yellow-100', text: 'text-yellow-800' },
    retired: { bg: 'bg-gray-100', text: 'text-gray-800' },
    compromised: { bg: 'bg-red-100', text: 'text-red-800' },
    valid: { bg: 'bg-green-100', text: 'text-green-800' },
    expired: { bg: 'bg-red-100', text: 'text-red-800' },
    revoked: { bg: 'bg-gray-100', text: 'text-gray-800' },
    pending: { bg: 'bg-blue-100', text: 'text-blue-800' }
  };

  const config = statusConfig[status] || { bg: 'bg-gray-100', text: 'text-gray-800' };

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </span>
  );
};

// Severity Badge Component
interface SeverityBadgeProps {
  severity: string;
}

const SeverityBadge: React.FC<SeverityBadgeProps> = ({ severity }) => {
  const severityConfig: Record<string, { bg: string; text: string; icon: React.ReactNode }> = {
    info: { 
      bg: 'bg-blue-100', 
      text: 'text-blue-800',
      icon: <Activity className="w-3 h-3 mr-1" />
    },
    warning: { 
      bg: 'bg-yellow-100', 
      text: 'text-yellow-800',
      icon: <AlertTriangle className="w-3 h-3 mr-1" />
    },
    error: { 
      bg: 'bg-red-100', 
      text: 'text-red-800',
      icon: <XCircle className="w-3 h-3 mr-1" />
    },
    critical: { 
      bg: 'bg-red-100', 
      text: 'text-red-800',
      icon: <XCircle className="w-3 h-3 mr-1" />
    }
  };

  const config = severityConfig[severity] || severityConfig.info;

  return (
    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${config.bg} ${config.text}`}>
      {config.icon}
      {severity.charAt(0).toUpperCase() + severity.slice(1)}
    </span>
  );
};

export default CryptoDashboard;
