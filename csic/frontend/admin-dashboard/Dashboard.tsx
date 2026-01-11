import React, { useState, useEffect } from 'react';
import {
  Users,
  Settings,
  Activity,
  Shield,
  Database,
  Server,
  Bell,
  LogOut,
  Search,
  ChevronDown,
  Menu,
  X,
  RefreshCw,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Clock,
  TrendingUp,
  TrendingDown,
  Download,
  Upload,
  Plus,
  Edit,
  Trash2,
  Eye,
  MoreHorizontal,
  Cpu,
  HardDrive,
  Wifi,
  Zap,
  Globe,
  Lock,
  Key,
  FileText,
  BarChart3,
  PieChart,
  Calendar,
  UserCheck,
  UserX,
  Mail,
  MessageSquare,
  Link,
  Unlink,
} from 'lucide-react';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart as RechartsPie,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';

// Types
interface User {
  id: string;
  name: string;
  email: string;
  role: 'admin' | 'operator' | 'analyst' | 'viewer';
  status: 'active' | 'inactive' | 'suspended';
  lastLogin: Date;
  mfaEnabled: boolean;
  department: string;
}

interface Service {
  id: string;
  name: string;
  status: 'healthy' | 'degraded' | 'down' | 'maintenance';
  uptime: number;
  responseTime: number;
  requestsPerSecond: number;
  errorRate: number;
  cpuUsage: number;
  memoryUsage: number;
  lastChecked: Date;
}

interface AuditLog {
  id: string;
  action: string;
  user: string;
  resource: string;
  ip: string;
  timestamp: Date;
  status: 'success' | 'failure';
  details: string;
}

interface SystemMetric {
  cpu: number;
  memory: number;
  disk: number;
  network: number;
  activeConnections: number;
  queueSize: number;
}

interface Alert {
  id: string;
  severity: 'critical' | 'high' | 'medium' | 'low';
  message: string;
  source: string;
  timestamp: Date;
  acknowledged: boolean;
}

// Mock Data
const mockUsers: User[] = [
  { id: '1', name: 'John Smith', email: 'john@csic.io', role: 'admin', status: 'active', lastLogin: new Date(), mfaEnabled: true, department: 'Security' },
  { id: '2', name: 'Sarah Johnson', email: 'sarah@csic.io', role: 'operator', status: 'active', lastLogin: new Date(Date.now() - 3600000), mfaEnabled: true, department: 'Operations' },
  { id: '3', name: 'Mike Chen', email: 'mike@csic.io', role: 'analyst', status: 'active', lastLogin: new Date(Date.now() - 7200000), mfaEnabled: false, department: 'Compliance' },
  { id: '4', name: 'Emily Davis', email: 'emily@csic.io', role: 'viewer', status: 'inactive', lastLogin: new Date(Date.now() - 86400000 * 7), mfaEnabled: true, department: 'Audit' },
  { id: '5', name: 'Robert Wilson', email: 'robert@csic.io', role: 'operator', status: 'suspended', lastLogin: new Date(Date.now() - 86400000 * 2), mfaEnabled: false, department: 'Operations' },
];

const mockServices: Service[] = [
  { id: '1', name: 'API Gateway', status: 'healthy', uptime: 99.99, responseTime: 45, requestsPerSecond: 1250, errorRate: 0.01, cpuUsage: 35, memoryUsage: 52, lastChecked: new Date() },
  { id: '2', name: 'Transaction Monitor', status: 'healthy', uptime: 99.95, responseTime: 120, requestsPerSecond: 450, errorRate: 0.05, cpuUsage: 68, memoryUsage: 71, lastChecked: new Date() },
  { id: '3', name: 'Compliance Engine', status: 'degraded', uptime: 99.80, responseTime: 350, requestsPerSecond: 180, errorRate: 0.15, cpuUsage: 82, memoryUsage: 78, lastChecked: new Date() },
  { id: '4', name: 'Risk Engine', status: 'healthy', uptime: 99.97, responseTime: 85, requestsPerSecond: 320, errorRate: 0.02, cpuUsage: 45, memoryUsage: 55, lastChecked: new Date() },
  { id: '5', name: 'Forensic Tools', status: 'healthy', uptime: 99.92, responseTime: 200, requestsPerSecond: 50, errorRate: 0.03, cpuUsage: 38, memoryUsage: 48, lastChecked: new Date() },
  { id: '6', name: 'Incident Response', status: 'maintenance', uptime: 98.50, responseTime: 150, requestsPerSecond: 25, errorRate: 0.10, cpuUsage: 25, memoryUsage: 35, lastChecked: new Date() },
  { id: '7', name: 'Oversight Service', status: 'healthy', uptime: 99.99, responseTime: 65, requestsPerSecond: 890, errorRate: 0.01, cpuUsage: 42, memoryUsage: 58, lastChecked: new Date() },
  { id: '8', name: 'Database', status: 'healthy', uptime: 99.999, responseTime: 15, requestsPerSecond: 2500, errorRate: 0.001, cpuUsage: 55, memoryUsage: 68, lastChecked: new Date() },
];

const mockAuditLogs: AuditLog[] = [
  { id: '1', action: 'USER_LOGIN', user: 'John Smith', resource: '/auth/login', ip: '192.168.1.100', timestamp: new Date(Date.now() - 300000), status: 'success', details: 'Successful login with MFA' },
  { id: '2', action: 'USER_LOGIN_FAILED', user: 'Robert Wilson', resource: '/auth/login', ip: '10.0.0.50', timestamp: new Date(Date.now() - 600000), status: 'failure', details: 'Invalid credentials - 3 attempts' },
  { id: '3', action: 'CONFIG_UPDATE', user: 'Sarah Johnson', resource: '/config/system', ip: '192.168.1.105', timestamp: new Date(Date.now() - 900000), status: 'success', details: 'Updated timeout settings' },
  { id: '4', action: 'USER_SUSPENDED', user: 'System', resource: '/users/5', ip: '127.0.0.1', timestamp: new Date(Date.now() - 1800000), status: 'success', details: 'Account suspended due to security policy' },
  { id: '5', action: 'EXPORT_DATA', user: 'Mike Chen', resource: '/reports/export', ip: '192.168.1.120', timestamp: new Date(Date.now() - 3600000), status: 'success', details: 'Exported compliance report' },
  { id: '6', action: 'API_KEY_CREATED', user: 'John Smith', resource: '/api/keys', ip: '192.168.1.100', timestamp: new Date(Date.now() - 7200000), status: 'success', details: 'New API key created for monitoring' },
  { id: '7', action: 'PERMISSION_DENIED', user: 'Emily Davis', resource: '/admin/users', ip: '192.168.1.130', timestamp: new Date(Date.now() - 10800000), status: 'failure', details: 'Insufficient permissions for action' },
  { id: '8', action: 'BULK_UPDATE', user: 'Sarah Johnson', resource: '/wallets/freeze', ip: '192.168.1.105', timestamp: new Date(Date.now() - 14400000), status: 'success', details: 'Frozen 5 wallets for compliance' },
];

const mockAlerts: Alert[] = [
  { id: '1', severity: 'critical', message: 'Database connection pool exhausted', source: 'Database', timestamp: new Date(Date.now() - 300000), acknowledged: false },
  { id: '2', severity: 'high', message: 'Compliance Engine response time degradation', source: 'Compliance Engine', timestamp: new Date(Date.now() - 900000), acknowledged: true },
  { id: '3', severity: 'medium', message: 'High memory usage on Transaction Monitor', source: 'Transaction Monitor', timestamp: new Date(Date.now() - 1800000), acknowledged: false },
  { id: '4', severity: 'low', message: 'Scheduled maintenance window starting soon', source: 'Incident Response', timestamp: new Date(Date.now() - 3600000), acknowledged: true },
  { id: '5', severity: 'high', message: 'Multiple failed login attempts detected', source: 'Auth Service', timestamp: new Date(Date.now() - 7200000), acknowledged: false },
];

// Store
interface AdminStore {
  currentView: string;
  setCurrentView: (view: string) => void;
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  users: User[];
  services: Service[];
  auditLogs: AuditLog[];
  alerts: Alert[];
  systemMetrics: SystemMetric;
  refreshData: () => void;
}

const useAdminStore = create<AdminStore>((set, get) => ({
  currentView: 'overview',
  setCurrentView: (view) => set({ currentView: view }),
  sidebarOpen: true,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  users: mockUsers,
  services: mockServices,
  auditLogs: mockAuditLogs,
  alerts: mockAlerts,
  systemMetrics: {
    cpu: 52,
    memory: 68,
    disk: 45,
    network: 78,
    activeConnections: 1250,
    queueSize: 15,
  },
  refreshData: () => {
    // Simulate data refresh
    set({
      systemMetrics: {
        cpu: 40 + Math.random() * 30,
        memory: 60 + Math.random() * 20,
        disk: 40 + Math.random() * 10,
        network: 60 + Math.random() * 30,
        activeConnections: 1000 + Math.floor(Math.random() * 500),
        queueSize: Math.floor(Math.random() * 30),
      },
    });
  },
}));

// Main Dashboard Component
export default function AdminDashboard() {
  const { currentView, setCurrentView, sidebarOpen, setSidebarOpen, alerts, refreshData } = useAdminStore();
  const [notificationsOpen, setNotificationsOpen] = useState(false);

  useEffect(() => {
    const interval = setInterval(refreshData, 10000);
    return () => clearInterval(interval);
  }, [refreshData]);

  const views = [
    { id: 'overview', label: 'Overview', icon: BarChart3 },
    { id: 'users', label: 'User Management', icon: Users },
    { id: 'services', label: 'Services', icon: Server },
    { id: 'audit', label: 'Audit Logs', icon: FileText },
    { id: 'security', label: 'Security', icon: Shield },
    { id: 'settings', label: 'Settings', icon: Settings },
  ];

  const unreadAlerts = alerts.filter(a => !a.acknowledged).length;

  const renderContent = () => {
    switch (currentView) {
      case 'overview': return <OverviewView />;
      case 'users': return <UsersView />;
      case 'services': return <ServicesView />;
      case 'audit': return <AuditView />;
      case 'security': return <SecurityView />;
      case 'settings': return <SettingsView />;
      default: return <OverviewView />;
    }
  };

  return (
    <div className="min-h-screen bg-gray-900 text-white flex">
      {/* Sidebar */}
      <aside className={`${sidebarOpen ? 'w-64' : 'w-20'} bg-gray-800 border-r border-gray-700 flex flex-col transition-all duration-300 fixed h-full z-30`}>
        <div className="h-16 flex items-center justify-between px-4 border-b border-gray-700">
          {sidebarOpen && (
            <div className="flex items-center gap-2">
              <div className="w-8 h-8 bg-primary-600 rounded-lg flex items-center justify-center">
                <Shield className="w-5 h-5 text-white" />
              </div>
              <span className="font-bold">CSIC Admin</span>
            </div>
          )}
          <button onClick={() => setSidebarOpen(!sidebarOpen)} className="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400">
            <Menu size={20} />
          </button>
        </div>

        <nav className="flex-1 py-4 px-3 space-y-1">
          {views.map((view) => (
            <button
              key={view.id}
              onClick={() => setCurrentView(view.id)}
              className={`w-full flex items-center gap-3 px-3 py-2.5 rounded-lg transition-colors ${
                currentView === view.id
                  ? 'bg-primary-600/20 text-primary-400 border border-primary-600/30'
                  : 'text-gray-400 hover:bg-gray-700/50 hover:text-white'
              }`}
            >
              <view.icon size={20} />
              {sidebarOpen && <span className="font-medium">{view.label}</span>}
            </button>
          ))}
        </nav>

        {sidebarOpen && (
          <div className="p-4 border-t border-gray-700">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center">
                <Users className="w-5 h-5 text-gray-400" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">Admin User</p>
                <p className="text-xs text-gray-500 truncate">admin@csic.io</p>
              </div>
            </div>
          </div>
        )}
      </aside>

      {/* Main Content */}
      <div className={`flex-1 flex flex-col ${sidebarOpen ? 'ml-64' : 'ml-20'} transition-all duration-300`}>
        {/* Header */}
        <header className="h-16 bg-gray-800 border-b border-gray-700 flex items-center justify-between px-6 sticky top-0 z-20">
          <h1 className="text-lg font-semibold">{views.find(v => v.id === currentView)?.label}</h1>
          <div className="flex items-center gap-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-500" />
              <input type="text" placeholder="Search..." className="w-64 pl-10 pr-4 py-2 bg-gray-900 border border-gray-700 rounded-lg text-sm" />
            </div>
            <button onClick={() => setNotificationsOpen(!notificationsOpen)} className="relative p-2 rounded-lg hover:bg-gray-700 text-gray-400">
              <Bell size={20} />
              {unreadAlerts > 0 && (
                <span className="absolute -top-1 -right-1 w-5 h-5 bg-red-500 rounded-full text-xs flex items-center justify-center">{unreadAlerts}</span>
              )}
            </button>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 p-6 overflow-auto">{renderContent()}</main>
      </div>

      {/* Notifications Panel */}
      {notificationsOpen && (
        <div className="fixed right-0 top-0 h-full w-80 bg-gray-800 border-l border-gray-700 shadow-xl z-50">
          <div className="p-4 border-b border-gray-700 flex items-center justify-between">
            <h3 className="font-semibold">Notifications</h3>
            <button onClick={() => setNotificationsOpen(false)} className="p-1 rounded hover:bg-gray-700"><X size={18} /></button>
          </div>
          <div className="overflow-y-auto h-full pb-20">
            {alerts.map(alert => (
              <div key={alert.id} className={`p-4 border-b border-gray-700 hover:bg-gray-700/50 ${!alert.acknowledged ? 'bg-primary-500/10' : ''}`}>
                <div className="flex items-start gap-3">
                  <AlertTriangle size={18} className={`${getSeverityColor(alert.severity)} mt-0.5`} />
                  <div>
                    <p className="text-sm font-medium">{alert.message}</p>
                    <p className="text-xs text-gray-500 mt-1">{alert.source} • {formatTime(alert.timestamp)}</p>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

// Helper Functions
function getSeverityColor(severity: string): string {
  const colors: Record<string, string> = {
    critical: 'text-red-500',
    high: 'text-orange-500',
    medium: 'text-yellow-500',
    low: 'text-green-500',
  };
  return colors[severity] || 'text-gray-500';
}

function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    healthy: 'text-green-500',
    degraded: 'text-yellow-500',
    down: 'text-red-500',
    maintenance: 'text-blue-500',
    active: 'text-green-500',
    inactive: 'text-gray-500',
    suspended: 'text-orange-500',
  };
  return colors[status] || 'text-gray-500';
}

function formatTime(date: Date): string {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);
  if (minutes < 1) return 'Just now';
  if (minutes < 60) return `${minutes}m ago`;
  if (hours < 24) return `${hours}h ago`;
  return `${days}d ago`;
}

// Overview View
function OverviewView() {
  const { systemMetrics, services, alerts } = useAdminStore();

  const cpuData = Array.from({ length: 24 }, (_, i) => ({
    time: `${i}:00`,
    cpu: 30 + Math.random() * 40,
    memory: 50 + Math.random() * 20,
  }));

  const serviceStatusData = [
    { name: 'Healthy', value: services.filter(s => s.status === 'healthy').length, color: '#10b981' },
    { name: 'Degraded', value: services.filter(s => s.status === 'degraded').length, color: '#f59e0b' },
    { name: 'Down', value: services.filter(s => s.status === 'down').length, color: '#ef4444' },
    { name: 'Maintenance', value: services.filter(s => s.status === 'maintenance').length, color: '#3b82f6' },
  ];

  return (
    <div className="space-y-6">
      {/* System Metrics Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <MetricCard title="CPU Usage" value={`${systemMetrics.cpu.toFixed(1)}%`} icon={Cpu} color="bg-blue-500" trend={-2.5} />
        <MetricCard title="Memory Usage" value={`${systemMetrics.memory.toFixed(1)}%`} icon={Activity} color="bg-purple-500" trend={1.2} />
        <MetricCard title="Active Connections" value={systemMetrics.activeConnections.toLocaleString()} icon={Link} color="bg-green-500" trend={5.8} />
        <MetricCard title="Queue Size" value={systemMetrics.queueSize.toString()} icon={Database} color="bg-orange-500" trend={-15} />
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">System Performance (24h)</h3>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={cpuData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                <XAxis dataKey="time" stroke="#9ca3af" />
                <YAxis stroke="#9ca3af" />
                <Tooltip contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151' }} />
                <Legend />
                <Area type="monotone" dataKey="cpu" stroke="#3b82f6" fill="#3b82f6" fillOpacity={0.2} name="CPU %" />
                <Area type="monotone" dataKey="memory" stroke="#8b5cf6" fill="#8b5cf6" fillOpacity={0.2} name="Memory %" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Service Status</h3>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <RechartsPie>
                <Pie data={serviceStatusData} cx="50%" cy="50%" innerRadius={50} outerRadius={80} paddingAngle={2} dataKey="value">
                  {serviceStatusData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151' }} />
              </RechartsPie>
            </ResponsiveContainer>
          </div>
          <div className="grid grid-cols-2 gap-2 mt-4">
            {serviceStatusData.map(item => (
              <div key={item.name} className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-400">{item.name}</span>
                <span className="text-sm text-white font-medium ml-auto">{item.value}</span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Services and Alerts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Service Health</h3>
          <div className="space-y-3">
            {services.slice(0, 5).map(service => (
              <div key={service.id} className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg">
                <div className="flex items-center gap-3">
                  <Server size={18} className={getStatusColor(service.status)} />
                  <div>
                    <p className="font-medium text-sm">{service.name}</p>
                    <p className="text-xs text-gray-400">{service.responseTime}ms response</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <span className="text-xs text-gray-400">{service.uptime}% uptime</span>
                  <span className={`px-2 py-0.5 rounded text-xs capitalize ${getStatusColor(service.status)}`}>{service.status}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Recent Alerts</h3>
          <div className="space-y-3">
            {alerts.map(alert => (
              <div key={alert.id} className={`p-3 rounded-lg border ${alert.acknowledged ? 'bg-gray-700/30 border-gray-700' : 'bg-primary-500/10 border-primary-500/30'}`}>
                <div className="flex items-start justify-between">
                  <div className="flex items-start gap-3">
                    <AlertTriangle size={16} className={getSeverityColor(alert.severity)} />
                    <div>
                      <p className="text-sm font-medium">{alert.message}</p>
                      <p className="text-xs text-gray-400 mt-1">{alert.source} • {formatTime(alert.timestamp)}</p>
                    </div>
                  </div>
                  <span className={`px-2 py-0.5 rounded text-xs capitalize ${getSeverityColor(alert.severity)} bg-opacity-20`}>{alert.severity}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

// Users View
function UsersView() {
  const { users } = useAdminStore();
  const [filter, setFilter] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');

  const filteredUsers = users.filter(user => {
    const matchesFilter = filter === 'all' || user.status === filter;
    const matchesSearch = user.name.toLowerCase().includes(searchTerm.toLowerCase()) || user.email.toLowerCase().includes(searchTerm.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <input type="text" placeholder="Search users..." value={searchTerm} onChange={(e) => setSearchTerm(e.target.value)} className="input w-64" />
          <select value={filter} onChange={(e) => setFilter(e.target.value)} className="select w-40">
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
            <option value="suspended">Suspended</option>
          </select>
        </div>
        <button className="btn btn-primary flex items-center gap-2">
          <Plus size={18} /> Add User
        </button>
      </div>

      <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-900/50">
            <tr>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">User</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Role</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Department</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Status</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">MFA</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Last Login</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {filteredUsers.map(user => (
              <tr key={user.id} className="hover:bg-gray-700/30">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center">
                      <Users size={18} className="text-gray-400" />
                    </div>
                    <div>
                      <p className="font-medium">{user.name}</p>
                      <p className="text-sm text-gray-400">{user.email}</p>
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  <span className="px-2 py-0.5 rounded text-xs bg-gray-700 text-gray-300 capitalize">{user.role}</span>
                </td>
                <td className="px-4 py-3 text-gray-400">{user.department}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-0.5 rounded text-xs capitalize ${getStatusColor(user.status)} bg-opacity-20`}>{user.status}</span>
                </td>
                <td className="px-4 py-3">
                  {user.mfaEnabled ? <CheckCircle size={18} className="text-green-500" /> : <XCircle size={18} className="text-gray-500" />}
                </td>
                <td className="px-4 py-3 text-gray-400 text-sm">{formatTime(user.lastLogin)}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <button className="p-1.5 rounded hover:bg-gray-700 text-gray-400"><Eye size={16} /></button>
                    <button className="p-1.5 rounded hover:bg-gray-700 text-gray-400"><Edit size={16} /></button>
                    <button className="p-1.5 rounded hover:bg-gray-700 text-red-400"><Trash2 size={16} /></button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// Services View
function ServicesView() {
  const { services, refreshData } = useAdminStore();

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm text-gray-400">Last updated: {new Date().toLocaleTimeString()}</span>
          <button onClick={refreshData} className="p-2 rounded-lg hover:bg-gray-700 text-gray-400">
            <RefreshCw size={16} />
          </button>
        </div>
        <div className="flex items-center gap-2">
          <span className="flex items-center gap-2 text-sm">
            <span className="w-2 h-2 rounded-full bg-green-500"></span> Healthy ({services.filter(s => s.status === 'healthy').length})
          </span>
          <span className="flex items-center gap-2 text-sm">
            <span className="w-2 h-2 rounded-full bg-yellow-500"></span> Degraded ({services.filter(s => s.status === 'degraded').length})
          </span>
          <span className="flex items-center gap-2 text-sm">
            <span className="w-2 h-2 rounded-full bg-red-500"></span> Down ({services.filter(s => s.status === 'down').length})
          </span>
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {services.map(service => (
          <div key={service.id} className="bg-gray-800 rounded-xl p-6 border border-gray-700">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center gap-3">
                <Server size={24} className={getStatusColor(service.status)} />
                <div>
                  <h3 className="font-semibold">{service.name}</h3>
                  <p className={`text-sm ${getStatusColor(service.status)}`}>{service.status}</p>
                </div>
              </div>
              <button className="p-1.5 rounded hover:bg-gray-700 text-gray-400"><MoreHorizontal size={18} /></button>
            </div>

            <div className="grid grid-cols-2 gap-4 mb-4">
              <div className="bg-gray-700/50 rounded-lg p-3">
                <p className="text-xs text-gray-400">Uptime</p>
                <p className="text-lg font-semibold">{service.uptime}%</p>
              </div>
              <div className="bg-gray-700/50 rounded-lg p-3">
                <p className="text-xs text-gray-400">Response Time</p>
                <p className="text-lg font-semibold">{service.responseTime}ms</p>
              </div>
              <div className="bg-gray-700/50 rounded-lg p-3">
                <p className="text-xs text-gray-400">Requests/sec</p>
                <p className="text-lg font-semibold">{service.requestsPerSecond}</p>
              </div>
              <div className="bg-gray-700/50 rounded-lg p-3">
                <p className="text-xs text-gray-400">Error Rate</p>
                <p className="text-lg font-semibold text-red-400">{service.errorRate}%</p>
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-400">CPU</span>
                <span>{service.cpuUsage}%</span>
              </div>
              <div className="w-full h-2 bg-gray-700 rounded-full overflow-hidden">
                <div className="h-full bg-blue-500 rounded-full" style={{ width: `${service.cpuUsage}%` }} />
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-gray-400">Memory</span>
                <span>{service.memoryUsage}%</span>
              </div>
              <div className="w-full h-2 bg-gray-700 rounded-full overflow-hidden">
                <div className="h-full bg-purple-500 rounded-full" style={{ width: `${service.memoryUsage}%` }} />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

// Audit View
function AuditView() {
  const { auditLogs } = useAdminStore();
  const [filter, setFilter] = useState('all');

  const filteredLogs = auditLogs.filter(log => filter === 'all' || log.status === filter);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <select value={filter} onChange={(e) => setFilter(e.target.value)} className="select w-40">
            <option value="all">All Status</option>
            <option value="success">Success</option>
            <option value="failure">Failure</option>
          </select>
          <button className="btn btn-secondary flex items-center gap-2">
            <Download size={18} /> Export
          </button>
        </div>
        <input type="text" placeholder="Search logs..." className="input w-64" />
      </div>

      <div className="bg-gray-800 rounded-xl border border-gray-700 overflow-hidden">
        <table className="w-full">
          <thead className="bg-gray-900/50">
            <tr>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Timestamp</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Action</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">User</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Resource</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">IP Address</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Status</th>
              <th className="px-4 py-3 text-left text-sm font-medium text-gray-400">Details</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-700">
            {filteredLogs.map(log => (
              <tr key={log.id} className="hover:bg-gray-700/30">
                <td className="px-4 py-3 text-sm text-gray-400">{formatTime(log.timestamp)}</td>
                <td className="px-4 py-3 font-medium">{log.action.replace(/_/g, ' ')}</td>
                <td className="px-4 py-3">{log.user}</td>
                <td className="px-4 py-3 text-gray-400 text-sm">{log.resource}</td>
                <td className="px-4 py-3 font-mono text-sm text-gray-400">{log.ip}</td>
                <td className="px-4 py-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${log.status === 'success' ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'}`}>
                    {log.status}
                  </span>
                </td>
                <td className="px-4 py-3 text-sm text-gray-400 max-w-xs truncate">{log.details}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// Security View
function SecurityView() {
  const { users } = useAdminStore();

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Security Overview</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-gray-400">MFA Compliance</span>
              <span className="text-green-500">{Math.round(users.filter(u => u.mfaEnabled).length / users.length * 100)}%</span>
            </div>
            <div className="w-full h-2 bg-gray-700 rounded-full overflow-hidden">
              <div className="h-full bg-green-500 rounded-full" style={{ width: `${users.filter(u => u.mfaEnabled).length / users.length * 100}%` }} />
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Active Users</span>
              <span>{users.filter(u => u.status === 'active').length}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Suspended Accounts</span>
              <span className="text-orange-500">{users.filter(u => u.status === 'suspended').length}</span>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">API Keys</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg">
              <div className="flex items-center gap-2">
                <Key size={18} className="text-gray-400" />
                <span className="font-mono text-sm">sk-...x7k9</span>
              </div>
              <span className="text-xs text-green-500">Active</span>
            </div>
            <div className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg">
              <div className="flex items-center gap-2">
                <Key size={18} className="text-gray-400" />
                <span className="font-mono text-sm">pk-...m2n5</span>
              </div>
              <span className="text-xs text-gray-500">Revoked</span>
            </div>
            <button className="w-full btn btn-secondary text-sm mt-2">Manage API Keys</button>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Encryption Status</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Data at Rest</span>
              <span className="flex items-center gap-1 text-green-500"><Lock size={14} /> AES-256</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Data in Transit</span>
              <span className="flex items-center gap-1 text-green-500"><Shield size={14} /> TLS 1.3</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Key Rotation</span>
              <span className="text-gray-300">30 days</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">HSM Status</span>
              <span className="flex items-center gap-1 text-green-500"><CheckCircle size={14} /> Connected</span>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
        <h3 className="text-lg font-semibold mb-4">Recent Security Events</h3>
        <div className="space-y-3">
          {[
            { event: 'Failed login attempt', user: 'Robert Wilson', time: '10 minutes ago', severity: 'high' },
            { event: 'API key created', user: 'John Smith', time: '2 hours ago', severity: 'low' },
            { event: 'Account suspended', user: 'Robert Wilson', time: '30 minutes ago', severity: 'high' },
            { event: 'Permission denied', user: 'Emily Davis', time: '3 hours ago', severity: 'medium' },
          ].map((item, index) => (
            <div key={index} className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg">
              <div className="flex items-center gap-3">
                <AlertTriangle size={18} className={getSeverityColor(item.severity)} />
                <div>
                  <p className="font-medium text-sm">{item.event}</p>
                  <p className="text-xs text-gray-400">{item.user} • {item.time}</p>
                </div>
              </div>
              <span className={`px-2 py-0.5 rounded text-xs capitalize ${getSeverityColor(item.severity)} bg-opacity-20`}>{item.severity}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// Settings View
function SettingsView() {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">General Settings</h3>
          <div className="space-y-4">
            <div>
              <label className="block text-sm text-gray-400 mb-2">Organization Name</label>
              <input type="text" defaultValue="CSIC Platform" className="input" />
            </div>
            <div>
              <label className="block text-sm text-gray-400 mb-2">Session Timeout (minutes)</label>
              <input type="number" defaultValue="30" className="input w-32" />
            </div>
            <div>
              <label className="block text-sm text-gray-400 mb-2">Default Timezone</label>
              <select className="select">
                <option>UTC</option>
                <option>America/New_York</option>
                <option>Europe/London</option>
                <option>Asia/Tokyo</option>
              </select>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Security Settings</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Require MFA</p>
                <p className="text-sm text-gray-400">All users must enable MFA</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" defaultChecked />
                <div className="w-11 h-6 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">IP Whitelisting</p>
                <p className="text-sm text-gray-400">Restrict access to specific IPs</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" />
                <div className="w-11 h-6 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Audit Logging</p>
                <p className="text-sm text-gray-400">Log all user actions</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" defaultChecked />
                <div className="w-11 h-6 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Notification Settings</h3>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Email Notifications</p>
                <p className="text-sm text-gray-400">Receive alerts via email</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" defaultChecked />
                <div className="w-11 h-6 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Critical Alerts</p>
                <p className="text-sm text-gray-400">Immediate notification for critical issues</p>
              </div>
              <label className="relative inline-flex items-center cursor-pointer">
                <input type="checkbox" className="sr-only peer" defaultChecked />
                <div className="w-11 h-6 bg-gray-700 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-primary-600"></div>
              </label>
            </div>
          </div>
        </div>

        <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
          <h3 className="text-lg font-semibold mb-4">Integrations</h3>
          <div className="space-y-3">
            {[
              { name: 'Slack', status: 'connected' },
              { name: 'PagerDuty', status: 'connected' },
              { name: 'Jira', status: 'disconnected' },
              { name: 'ServiceNow', status: 'disconnected' },
            ].map((integration, index) => (
              <div key={index} className="flex items-center justify-between p-3 bg-gray-700/50 rounded-lg">
                <span className="font-medium">{integration.name}</span>
                <span className={`text-sm ${integration.status === 'connected' ? 'text-green-500' : 'text-gray-500'}`}>
                  {integration.status}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <button className="btn btn-primary">Save Changes</button>
      </div>
    </div>
  );
}

// Metric Card Component
function MetricCard({ title, value, icon: Icon, color, trend }: { title: string; value: string; icon: React.ElementType; color: string; trend?: number }) {
  return (
    <div className="bg-gray-800 rounded-xl p-6 border border-gray-700">
      <div className="flex items-center justify-between mb-4">
        <div className={`p-3 rounded-lg ${color}`}>
          <Icon className="w-6 h-6 text-white" />
        </div>
        {trend !== undefined && (
          <div className={`flex items-center gap-1 text-sm ${trend >= 0 ? 'text-green-400' : 'text-red-400'}`}>
            {trend >= 0 ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
            {Math.abs(trend)}%
          </div>
        )}
      </div>
      <h3 className="text-gray-400 text-sm">{title}</h3>
      <p className="text-2xl font-bold text-white mt-1">{value}</p>
    </div>
  );
}

// Import create from zustand
import { create } from 'zustand';
