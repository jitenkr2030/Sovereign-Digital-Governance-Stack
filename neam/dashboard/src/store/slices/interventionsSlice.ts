/**
 * NEAM Dashboard - Interventions Slice
 * State management for policy interventions
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { Intervention, InterventionStatus, InterventionType } from '../../types';

interface InterventionsState {
  interventions: Intervention[];
  selectedIntervention: Intervention | null;
  filters: {
    status?: InterventionStatus[];
    type?: InterventionType[];
    region?: string;
  };
  isLoading: boolean;
  error: string | null;
}

const initialState: InterventionsState = {
  interventions: [],
  selectedIntervention: null,
  filters: {},
  isLoading: false,
  error: null,
};

const interventionsSlice = createSlice({
  name: 'interventions',
  initialState,
  reducers: {
    setInterventions: (state, action: PayloadAction<Intervention[]>) => {
      state.interventions = action.payload;
      state.isLoading = false;
    },
    addIntervention: (state, action: PayloadAction<Intervention>) => {
      state.interventions = [action.payload, ...state.interventions];
    },
    updateIntervention: (state, action: PayloadAction<Intervention>) => {
      const index = state.interventions.findIndex((i) => i.id === action.payload.id);
      if (index !== -1) {
        state.interventions[index] = action.payload;
      }
      if (state.selectedIntervention?.id === action.payload.id) {
        state.selectedIntervention = action.payload;
      }
    },
    removeIntervention: (state, action: PayloadAction<string>) => {
      state.interventions = state.interventions.filter((i) => i.id !== action.payload);
    },
    setSelectedIntervention: (state, action: PayloadAction<Intervention | null>) => {
      state.selectedIntervention = action.payload;
    },
    setFilters: (state, action: PayloadAction<InterventionsState['filters']>) => {
      state.filters = { ...state.filters, ...action.payload };
    },
    clearFilters: (state) => {
      state.filters = {};
    },
    setLoading: (state, action: PayloadAction<boolean>) => {
      state.isLoading = action.payload;
    },
    setError: (state, action: PayloadAction<string | null>) => {
      state.error = action.payload;
    },
    updateProgress: (state, action: PayloadAction<{ id: string; progress: number }>) => {
      const intervention = state.interventions.find((i) => i.id === action.payload.id);
      if (intervention) {
        intervention.progress = action.payload.progress;
      }
    },
    updateStatus: (state, action: PayloadAction<{ id: string; status: InterventionStatus }>) => {
      const intervention = state.interventions.find((i) => i.id === action.payload.id);
      if (intervention) {
        intervention.status = action.payload.status;
      }
    },
  },
});

export const {
  setInterventions,
  addIntervention,
  updateIntervention,
  removeIntervention,
  setSelectedIntervention,
  setFilters,
  clearFilters,
  setLoading,
  setError,
  updateProgress,
  updateStatus,
} = interventionsSlice.actions;

export default interventionsSlice.reducer;

// Selectors
export const selectInterventions = (state: { interventions: InterventionsState }) =>
  state.interventions.interventions;

export const selectActiveInterventions = (state: { interventions: InterventionsState }) =>
  state.interventions.interventions.filter((i) =>
    ['APPROVED', 'EXECUTING', 'IN_PROGRESS'].includes(i.status)
  );

export const selectInterventionById = (id: string) => (state: { interventions: InterventionsState }) =>
  state.interventions.interventions.find((i) => i.id === id);

export const selectInterventionsByRegion = (regionId: string) => (state: { interventions: InterventionsState }) =>
  state.interventions.interventions.filter((i) => i.targetRegions.includes(regionId));

export const selectInterventionStats = (state: { interventions: InterventionsState }) => {
  const interventions = state.interventions.interventions;
  return {
    total: interventions.length,
    active: interventions.filter((i) => i.status === 'EXECUTING' || i.status === 'IN_PROGRESS').length,
    completed: interventions.filter((i) => i.status === 'COMPLETED').length,
    pending: interventions.filter((i) => i.status === 'PENDING').length,
    onHold: interventions.filter((i) => i.status === 'ON_HOLD').length,
  };
};
