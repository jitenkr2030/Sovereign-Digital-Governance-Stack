/**
 * NEAM Dashboard - UI Slice
 * UI state management for the dashboard
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Notification } from '../../types';

interface UIState {
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  loading: Record<string, boolean>;
  modals: Record<string, boolean>;
  toasts: Notification[];
  activeView: 'overview' | 'analytics' | 'reports' | 'interventions' | 'settings';
  sidebarView: 'standard' | 'minimized';
  globalSearch: string;
  showDrillDown: boolean;
  drillDownHistory: string[];
}

const initialState: UIState = {
  sidebarCollapsed: false,
  theme: 'light',
  loading: {},
  modals: {},
  toasts: [],
  activeView: 'overview',
  sidebarView: 'standard',
  globalSearch: '',
  showDrillDown: false,
  drillDownHistory: [],
};

const uiSlice = createSlice({
  name: 'ui',
  initialState,
  reducers: {
    toggleSidebar: (state) => {
      state.sidebarCollapsed = !state.sidebarCollapsed;
    },
    setSidebarCollapsed: (state, action: PayloadAction<boolean>) => {
      state.sidebarCollapsed = action.payload;
    },
    toggleTheme: (state) => {
      state.theme = state.theme === 'light' ? 'dark' : 'light';
    },
    setTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
      state.theme = action.payload;
    },
    setLoading: (state, action: PayloadAction<{ key: string; value: boolean }>) => {
      state.loading[action.payload.key] = action.payload.value;
    },
    setActiveView: (state, action: PayloadAction<UIState['activeView']>) => {
      state.activeView = action.payload;
    },
    setSidebarView: (state, action: PayloadAction<UIState['sidebarView']>) => {
      state.sidebarView = action.payload;
    },
    openModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = true;
    },
    closeModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = false;
    },
    toggleModal: (state, action: PayloadAction<string>) => {
      state.modals[action.payload] = !state.modals[action.payload];
    },
    addToast: (state, action: PayloadAction<Notification>) => {
      state.toasts = [action.payload, ...state.toasts].slice(0, 10); // Keep max 10 toasts
    },
    removeToast: (state, action: PayloadAction<string>) => {
      state.toasts = state.toasts.filter((t) => t.id !== action.payload);
    },
    clearToasts: (state) => {
      state.toasts = [];
    },
    setGlobalSearch: (state, action: PayloadAction<string>) => {
      state.globalSearch = action.payload;
    },
    clearGlobalSearch: (state) => {
      state.globalSearch = '';
    },
    setShowDrillDown: (state, action: PayloadAction<boolean>) => {
      state.showDrillDown = action.payload;
    },
    pushDrillDownHistory: (state, action: PayloadAction<string>) => {
      state.drillDownHistory.push(action.payload);
    },
    popDrillDownHistory: (state) => {
      state.drillDownHistory.pop();
    },
    clearDrillDownHistory: (state) => {
      state.drillDownHistory = [];
    },
    resetUI: () => initialState,
  },
});

export const {
  toggleSidebar,
  setSidebarCollapsed,
  toggleTheme,
  setTheme,
  setLoading,
  setActiveView,
  setSidebarView,
  openModal,
  closeModal,
  toggleModal,
  addToast,
  removeToast,
  clearToasts,
  setGlobalSearch,
  clearGlobalSearch,
  setShowDrillDown,
  pushDrillDownHistory,
  popDrillDownHistory,
  clearDrillDownHistory,
  resetUI,
} = uiSlice.actions;

export default uiSlice.reducer;

// Selectors
export const selectSidebarCollapsed = (state: { ui: UIState }) => state.ui.sidebarCollapsed;
export const selectTheme = (state: { ui: UIState }) => state.ui.theme;
export const selectActiveView = (state: { ui: UIState }) => state.ui.activeView;
export const selectToasts = (state: { ui: UIState }) => state.ui.toasts;
export const selectIsLoading = (key: string) => (state: { ui: UIState }) => state.ui.loading[key] || false;
export const selectIsModalOpen = (modalId: string) => (state: { ui: UIState }) => state.ui.modals[modalId] || false;
