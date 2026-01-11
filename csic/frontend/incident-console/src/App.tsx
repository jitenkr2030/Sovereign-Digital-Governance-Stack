import React, { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { useIncidentStore } from '../store/incidentStore';
import Layout from '../components/Layout';
import Dashboard from '../pages/Dashboard';
import IncidentList from '../pages/IncidentList';
import IncidentDetail from '../pages/IncidentDetail';
import CreateIncident from '../pages/CreateIncident';
import Analytics from '../pages/Analytics';
import Settings from '../pages/Settings';

const App: React.FC = () => {
  const { fetchIncidents, fetchDashboard, fetchNotifications } = useIncidentStore();

  useEffect(() => {
    // Initialize data
    fetchIncidents();
    fetchDashboard();
    fetchNotifications();
  }, [fetchIncidents, fetchDashboard, fetchNotifications]);

  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<Navigate to="/dashboard" replace />} />
        <Route path="dashboard" element={<Dashboard />} />
        <Route path="incidents" element={<IncidentList />} />
        <Route path="incidents/new" element={<CreateIncident />} />
        <Route path="incidents/:id" element={<IncidentDetail />} />
        <Route path="analytics" element={<Analytics />} />
        <Route path="settings" element={<Settings />} />
      </Route>
    </Routes>
  );
};

export default App;
