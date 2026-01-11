import { create } from 'zustand';
import { Incident, IncidentFilter, DashboardData, Notification, IncidentStats } from '../types/incident';

interface IncidentState {
  // Incidents
  incidents: Incident[];
  selectedIncident: Incident | null;
  isLoading: boolean;
  error: string | null;
  
  // Filters
  filters: IncidentFilter;
  
  // Dashboard
  dashboardData: DashboardData | null;
  
  // Notifications
  notifications: Notification[];
  unreadCount: number;
  
  // Actions
  fetchIncidents: (filter?: IncidentFilter) => Promise<void>;
  fetchIncident: (id: string) => Promise<Incident | null>;
  createIncident: (incident: Partial<Incident>) => Promise<Incident>;
  updateIncident: (id: string, updates: Partial<Incident>) => Promise<Incident>;
  addTimelineEvent: (incidentId: string, event: Partial<Incident>) => Promise<void>;
  assignIncident: (incidentId: string, assigneeId: string) => Promise<void>;
  resolveIncident: (incidentId: string, resolution: string) => Promise<void>;
  
  // Dashboard
  fetchDashboard: () => Promise<void>;
  
  // Filters
  setFilters: (filters: IncidentFilter) => void;
  clearFilters: () => void;
  
  // Notifications
  fetchNotifications: () => Promise<void>;
  markNotificationRead: (id: string) => Promise<void>;
  markAllNotificationsRead: () => Promise<void>;
  
  // Selection
  selectIncident: (incident: Incident | null) => void;
}

export const useIncidentStore = create<IncidentState>((set, get) => ({
  // Initial State
  incidents: [],
  selectedIncident: null,
  isLoading: false,
  error: null,
  filters: {},
  dashboardData: null,
  notifications: [],
  unreadCount: 0,
  
  // Actions
  fetchIncidents: async (filter?: IncidentFilter) => {
    set({ isLoading: true, error: null });
    try {
      // Simulated API call - replace with actual API
      const mockIncidents: Incident[] = [
        {
          id: 'INC-001',
          title: 'Suspicious Transaction Pattern Detected',
          description: 'Multiple high-value transactions detected from previously flagged wallet addresses',
          severity: 'critical',
          status: 'investigating',
          type: 'fraud',
          priority: 1,
          reporter: { id: '1', name: 'John Doe', email: 'john@csic.io', role: 'Analyst', department: 'Security' },
          affectedSystems: ['Transaction Monitoring', 'Compliance DB'],
          indicators: [],
          timeline: [],
          relatedIncidents: [],
          tags: ['fraud', 'aml', 'suspicious'],
          createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
          updatedAt: new Date(),
          metrics: { timeToDetect: 15, timeToRespond: 5, timeToResolve: 0, affectedUsers: 0, estimatedImpact: 50000, recoveryCost: 0 },
        },
        {
          id: 'INC-002',
          title: 'Unauthorized Access Attempt',
          description: 'Multiple failed login attempts from IP addresses in restricted regions',
          severity: 'high',
          status: 'containment',
          type: 'security',
          priority: 2,
          reporter: { id: '2', name: 'Jane Smith', email: 'jane@csic.io', role: 'Security Analyst', department: 'Security' },
          affectedSystems: ['Authentication Service'],
          indicators: [],
          timeline: [],
          relatedIncidents: [],
          tags: ['security', 'brute-force', 'geo-restriction'],
          createdAt: new Date(Date.now() - 24 * 60 * 60 * 1000),
          updatedAt: new Date(),
          metrics: { timeToDetect: 2, timeToRespond: 8, timeToResolve: 0, affectedUsers: 5, estimatedImpact: 10000, recoveryCost: 0 },
        },
        {
          id: 'INC-003',
          title: 'Compliance Report Submission Delayed',
          description: 'Monthly regulatory filing deadline at risk due to data pipeline issue',
          severity: 'medium',
          status: 'open',
          type: 'compliance',
          priority: 3,
          reporter: { id: '3', name: 'Mike Johnson', email: 'mike@csic.io', role: 'Compliance Officer', department: 'Compliance' },
          affectedSystems: ['Reporting Service'],
          indicators: [],
          timeline: [],
          relatedIncidents: [],
          tags: ['compliance', 'regulatory', 'deadline'],
          createdAt: new Date(Date.now() - 4 * 60 * 60 * 1000),
          updatedAt: new Date(),
          metrics: { timeToDetect: 30, timeToRespond: 0, timeToResolve: 0, affectedUsers: 0, estimatedImpact: 25000, recoveryCost: 0 },
        },
      ];
      
      set({ incidents: mockIncidents, isLoading: false });
    } catch (error) {
      set({ error: 'Failed to fetch incidents', isLoading: false });
    }
  },
  
  fetchIncident: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      const { incidents } = get();
      const incident = incidents.find(i => i.id === id) || null;
      set({ selectedIncident: incident, isLoading: false });
      return incident;
    } catch (error) {
      set({ error: 'Failed to fetch incident', isLoading: false });
      return null;
    }
  },
  
  createIncident: async (incidentData: Partial<Incident>) => {
    set({ isLoading: true, error: null });
    try {
      const newIncident: Incident = {
        id: `INC-${String(get().incidents.length + 1).padStart(3, '0')}`,
        title: incidentData.title || '',
        description: incidentData.description || '',
        severity: incidentData.severity || 'medium',
        status: 'open',
        type: incidentData.type || 'technical',
        priority: incidentData.priority || 3,
        reporter: incidentData.reporter!,
        affectedSystems: incidentData.affectedSystems || [],
        indicators: [],
        timeline: [],
        relatedIncidents: [],
        tags: incidentData.tags || [],
        createdAt: new Date(),
        updatedAt: new Date(),
        metrics: { timeToDetect: 0, timeToRespond: 0, timeToResolve: 0, affectedUsers: 0, estimatedImpact: 0, recoveryCost: 0 },
      };
      
      set(state => ({
        incidents: [newIncident, ...state.incidents],
        isLoading: false,
      }));
      
      return newIncident;
    } catch (error) {
      set({ error: 'Failed to create incident', isLoading: false });
      throw error;
    }
  },
  
  updateIncident: async (id: string, updates: Partial<Incident>) => {
    set({ isLoading: true, error: null });
    try {
      set(state => ({
        incidents: state.incidents.map(i =>
          i.id === id ? { ...i, ...updates, updatedAt: new Date() } : i
        ),
        selectedIncident: state.selectedIncident?.id === id
          ? { ...state.selectedIncident, ...updates, updatedAt: new Date() }
          : state.selectedIncident,
        isLoading: false,
      }));
      
      const { incidents } = get();
      return incidents.find(i => i.id === id)!;
    } catch (error) {
      set({ error: 'Failed to update incident', isLoading: false });
      throw error;
    }
  },
  
  addTimelineEvent: async (incidentId: string, event: Partial<Incident>) => {
    // Implementation would add timeline event
  },
  
  assignIncident: async (incidentId: string, assigneeId: string) => {
    await get().updateIncident(incidentId, { assignee: { id: assigneeId, name: 'Assigned User', email: '', role: '', department: '' } });
  },
  
  resolveIncident: async (incidentId: string, resolution: string) => {
    await get().updateIncident(incidentId, { status: 'resolved', resolvedAt: new Date() });
  },
  
  fetchDashboard: async () => {
    set({ isLoading: true, error: null });
    try {
      const mockDashboardData: DashboardData = {
        stats: {
          total: 156,
          bySeverity: { critical: 3, high: 12, medium: 28, low: 45, info: 68 },
          byStatus: { open: 45, investigating: 32, containment: 15, resolved: 52, closed: 12 },
          byType: { security: 45, compliance: 32, technical: 28, operational: 35, fraud: 16 },
          averageResolutionTime: 180,
          openIncidentsCount: 45,
          criticalIncidentsCount: 3,
        },
        recentIncidents: [],
        trendData: [],
        topAffectedSystems: [],
        teamWorkload: [],
      };
      
      set({ dashboardData: mockDashboardData, isLoading: false });
    } catch (error) {
      set({ error: 'Failed to fetch dashboard', isLoading: false });
    }
  },
  
  setFilters: (filters: IncidentFilter) => {
    set({ filters });
  },
  
  clearFilters: () => {
    set({ filters: {} });
  },
  
  fetchNotifications: async () => {
    const mockNotifications: Notification[] = [
      {
        id: '1',
        type: 'incident_escalated',
        incidentId: 'INC-001',
        incidentTitle: 'Suspicious Transaction Pattern Detected',
        message: 'Incident INC-001 has been escalated to critical priority',
        read: false,
        createdAt: new Date(Date.now() - 30 * 60 * 1000),
      },
      {
        id: '2',
        type: 'comment_added',
        incidentId: 'INC-002',
        incidentTitle: 'Unauthorized Access Attempt',
        message: 'New comment added to incident INC-002',
        read: false,
        createdAt: new Date(Date.now() - 2 * 60 * 60 * 1000),
      },
    ];
    
    set({
      notifications: mockNotifications,
      unreadCount: mockNotifications.filter(n => !n.read).length,
    });
  },
  
  markNotificationRead: async (id: string) => {
    set(state => ({
      notifications: state.notifications.map(n =>
        n.id === id ? { ...n, read: true } : n
      ),
      unreadCount: Math.max(0, state.unreadCount - 1),
    }));
  },
  
  markAllNotificationsRead: async () => {
    set(state => ({
      notifications: state.notifications.map(n => ({ ...n, read: true })),
      unreadCount: 0,
    }));
  },
  
  selectIncident: (incident: Incident | null) => {
    set({ selectedIncident: incident });
  },
}));
