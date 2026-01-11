import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Alert, AlertType } from '../../types';

interface AlertState {
  alerts: Alert[];
  unreadCount: number;
  filters: {
    types?: AlertType[];
    priority?: string[];
    status?: string;
  };
  selectedAlert: Alert | null;
  isLoading: boolean;
  error: string | null;
}

const initialState: AlertState = {
  alerts: [],
  unreadCount: 0,
  filters: {},
  selectedAlert: null,
  isLoading: false,
  error: null,
};

const alertSlice = createSlice({
  name: 'alerts',
  initialState,
  reducers: {
    setAlerts: (state, action: PayloadAction<Alert[]>) => {
      state.alerts = action.payload;
      state.unreadCount = action.payload.filter((a) => !a.acknowledgedAt).length;
    },
    addAlert: (state, action: PayloadAction<Alert>) => {
      state.alerts = [action.payload, ...state.alerts];
      if (!action.payload.acknowledgedAt) {
        state.unreadCount++;
      }
    },
    updateAlert: (state, action: PayloadAction<Alert>) => {
      const index = state.alerts.findIndex((a) => a.id === action.payload.id);
      if (index !== -1) {
        state.alerts[index] = action.payload;
        state.unreadCount = state.alerts.filter((a) => !a.acknowledgedAt).length;
      }
    },
    removeAlert: (state, action: PayloadAction<string>) => {
      state.alerts = state.alerts.filter((a) => a.id !== action.payload);
      state.unreadCount = state.alerts.filter((a) => !a.acknowledgedAt).length;
    },
    acknowledgeAlert: (state, action: PayloadAction<string>) => {
      const alert = state.alerts.find((a) => a.id === action.payload);
      if (alert && !alert.acknowledgedAt) {
        alert.acknowledgedAt = new Date();
        state.unreadCount = Math.max(0, state.unreadCount - 1);
      }
    },
    setFilters: (state, action: PayloadAction<AlertState['filters']>) => {
      state.filters = { ...state.filters, ...action.payload };
    },
    clearFilters: (state) => {
      state.filters = {};
    },
    setSelectedAlert: (state, action: PayloadAction<Alert | null>) => {
      state.selectedAlert = action.payload;
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
  },
});

export const {
  setAlerts,
  addAlert,
  updateAlert,
  removeAlert,
  acknowledgeAlert,
  setFilters,
  clearFilters,
  setSelectedAlert,
  setLoading,
  setError,
} = alertSlice.actions;

export default alertSlice.reducer;
