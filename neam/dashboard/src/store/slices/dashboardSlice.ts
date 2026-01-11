import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { DashboardSummary, MetricCard, RegionData, HeatmapCell, DashboardFilters } from '../../types';

interface DashboardState {
  summary: DashboardSummary | null;
  metrics: MetricCard[];
  regions: RegionData[];
  heatmapData: HeatmapCell[];
  filters: DashboardFilters;
  selectedRegion: RegionData | null;
  viewMode: 'overview' | 'regional' | 'district';
  dateRange: { start: Date | null; end: Date | null };
  isLoading: boolean;
  error: string | null;
}

const initialState: DashboardState = {
  summary: null,
  metrics: [],
  regions: [],
  heatmapData: [],
  filters: {},
  selectedRegion: null,
  viewMode: 'overview',
  dateRange: { start: null, end: null },
  isLoading: false,
  error: null,
};

const dashboardSlice = createSlice({
  name: 'dashboard',
  initialState,
  reducers: {
    setSummary: (state, action: PayloadAction<DashboardSummary>) => {
      state.summary = action.payload;
    },
    setMetrics: (state, action: PayloadAction<MetricCard[]>) => {
      state.metrics = action.payload;
    },
    setRegions: (state, action: PayloadAction<RegionData[]>) => {
      state.regions = action.payload;
    },
    setHeatmapData: (state, action: PayloadAction<HeatmapCell[]>) => {
      state.heatmapData = action.payload;
    },
    setFilters: (state, action: PayloadAction<DashboardFilters>) => {
      state.filters = { ...state.filters, ...action.payload };
    },
    clearFilters: (state) => {
      state.filters = {};
    },
    setSelectedRegion: (state, action: PayloadAction<RegionData | null>) => {
      state.selectedRegion = action.payload;
    },
    setViewMode: (state, action: PayloadAction<'overview' | 'regional' | 'district'>) => {
      state.viewMode = action.payload;
    },
    setDateRange: (state, action: PayloadAction<{ start: Date | null; end: Date | null }>) => {
      state.dateRange = action.payload;
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
  setSummary,
  setMetrics,
  setRegions,
  setHeatmapData,
  setFilters,
  clearFilters,
  setSelectedRegion,
  setViewMode,
  setDateRange,
  setLoading,
  setError,
} = dashboardSlice.actions;

export default dashboardSlice.reducer;
