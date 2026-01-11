/**
 * NEAM Dashboard - Filters Slice
 * Global filter state management for dashboard
 */

import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { DashboardFilters, DateRange, DatePreset } from '../../types';

interface FiltersState {
  filters: DashboardFilters;
  dateRange: DateRange;
  preset: DatePreset | null;
  isCustomRange: boolean;
  appliedFilters: DashboardFilters;
}

const initialState: FiltersState = {
  filters: {},
  dateRange: {
    start: new Date(new Date().setMonth(new Date().getMonth() - 1)),
    end: new Date(),
    preset: 'MONTH',
  },
  preset: 'MONTH',
  isCustomRange: false,
  appliedFilters: {},
};

const filtersSlice = createSlice({
  name: 'filters',
  initialState,
  reducers: {
    setDateRange: (state, action: PayloadAction<DateRange>) => {
      state.dateRange = action.payload;
      state.preset = action.payload.preset || null;
      state.isCustomRange = !action.payload.preset;
    },
    setPreset: (state, action: PayloadAction<DatePreset>) => {
      const now = new Date();
      let start: Date;
      const end = new Date();

      switch (action.payload) {
        case 'TODAY':
          start = new Date(now.setHours(0, 0, 0, 0));
          break;
        case 'WEEK':
          start = new Date(now.setDate(now.getDate() - 7));
          break;
        case 'MONTH':
          start = new Date(now.setMonth(now.getMonth() - 1));
          break;
        case 'QUARTER':
          start = new Date(now.setMonth(now.getMonth() - 3));
          break;
        case 'YEAR':
          start = new Date(now.setFullYear(now.getFullYear() - 1));
          break;
        case 'FY':
          const currentMonth = new Date().getMonth();
          const currentFYStart = currentMonth >= 3 ? new Date(new Date().getFullYear(), 3, 1) : new Date(new Date().getFullYear() - 1, 3, 1);
          start = currentFYStart;
          break;
        default:
          start = new Date(now.setMonth(now.getMonth() - 1));
      }

      state.dateRange = { start, end, preset: action.payload };
      state.preset = action.payload;
      state.isCustomRange = false;
    },
    setCustomDateRange: (state, action: PayloadAction<{ start: Date; end: Date }>) => {
      state.dateRange = { ...action.payload, preset: undefined };
      state.preset = null;
      state.isCustomRange = true;
    },
    setRegions: (state, action: PayloadAction<string[]>) => {
      state.filters.regions = action.payload;
    },
    toggleRegion: (state, action: PayloadAction<string>) => {
      if (!state.filters.regions) {
        state.filters.regions = [];
      }
      const index = state.filters.regions.indexOf(action.payload);
      if (index === -1) {
        state.filters.regions.push(action.payload);
      } else {
        state.filters.regions.splice(index, 1);
      }
    },
    setSectors: (state, action: PayloadAction<string[]>) => {
      state.filters.sectors = action.payload;
    },
    toggleSector: (state, action: PayloadAction<string>) => {
      if (!state.filters.sectors) {
        state.filters.sectors = [];
      }
      const index = state.filters.sectors.indexOf(action.payload);
      if (index === -1) {
        state.filters.sectors.push(action.payload);
      } else {
        state.filters.sectors.splice(index, 1);
      }
    },
    setIndicators: (state, action: PayloadAction<string[]>) => {
      state.filters.indicators = action.payload;
    },
    setAlertTypes: (state, action: PayloadAction<string[]>) => {
      state.filters.alertTypes = action.payload as string[];
    },
    setComparisonMode: (state, action: PayloadAction<'period' | 'region' | 'sector'>) => {
      state.filters.comparisonMode = action.payload;
    },
    clearAllFilters: (state) => {
      state.filters = {};
      state.appliedFilters = {};
    },
    applyFilters: (state) => {
      state.appliedFilters = { ...state.filters };
    },
    resetToDefault: (state) => {
      state.filters = {};
      state.dateRange = initialState.dateRange;
      state.preset = initialState.preset;
      state.isCustomRange = false;
      state.appliedFilters = {};
    },
  },
});

export const {
  setDateRange,
  setPreset,
  setCustomDateRange,
  setRegions,
  toggleRegion,
  setSectors,
  toggleSector,
  setIndicators,
  setAlertTypes,
  setComparisonMode,
  clearAllFilters,
  applyFilters,
  resetToDefault,
} = filtersSlice.actions;

export default filtersSlice.reducer;

// Selectors
export const selectDateRange = (state: { filters: FiltersState }) => state.filters.dateRange;
export const selectActiveFilters = (state: { filters: FiltersState }) => state.filters.filters;
export const selectAppliedFilters = (state: { filters: FiltersState }) => state.filters.appliedFilters;
export const selectFilterCount = (state: { filters: FiltersState }) => {
  const f = state.filters.filters;
  return (f.regions?.length || 0) + (f.sectors?.length || 0) + (f.indicators?.length || 0);
};
