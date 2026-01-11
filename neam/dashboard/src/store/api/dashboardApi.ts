import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type {
  DashboardSummary,
  MetricCard,
  RegionData,
  HeatmapCell,
  AnalyticsData,
  DateRange,
} from '../../types';

export const dashboardApi = createApi({
  reducerPath: 'dashboardApi',
  baseQuery: fetchBaseQuery({
    baseUrl: import.meta.env.VITE_API_URL || '/api/v1',
    prepareHeaders: (headers) => {
      const token = localStorage.getItem('neam_token');
      if (token) {
        headers.set('Authorization', `Bearer ${token}`);
      }
      return headers;
    },
  }),
  endpoints: (builder) => ({
    getDashboardSummary: builder.query<DashboardSummary, void>({
      query: () => '/dashboard/summary',
    }),
    getMetricCards: builder.query<MetricCard[], void>({
      query: () => '/dashboard/metrics',
    }),
    getMetricTrend: builder.query<{ metricId: string; data: { timestamp: string; value: number }[] }, { metricId: string; days: number }>({
      query: ({ metricId, days }) => `/dashboard/metrics/${metricId}/trend?days=${days}`,
    }),
    getRegions: builder.query<RegionData[], { status?: string }>({
      query: (params) => `/dashboard/regions${params.status ? `?status=${params.status}` : ''}`,
    }),
    getRegion: builder.query<RegionData, string>({
      query: (regionId) => `/dashboard/regions/${regionId}`,
    }),
    getHeatmapData: builder.query<HeatmapCell[], { metric: string; date?: string }>({
      query: (params) => `/dashboard/heatmap?metric=${params.metric}${params.date ? `&date=${params.date}` : ''}`,
    }),
    getAnalytics: builder.query<AnalyticsData, { period: string; startDate: string; endDate: string }>({
      query: (params) => `/dashboard/analytics?period=${params.period}&start=${params.startDate}&end=${params.endDate}`,
    }),
    getComparison: builder.query<{ regions: { id: string; name: string; value: number }[] }, { metric: string; regions: string[] }>({
      query: (params) => {
        const regionParams = params.regions.join(',');
        return `/dashboard/compare?metric=${params.metric}&regions=${regionParams}`;
      },
    }),
    exportDashboard: builder.mutation<Blob, { format: string; sections: string[]; dateRange: DateRange }>({
      query: (params) => ({
        url: '/dashboard/export',
        method: 'POST',
        body: params,
        responseHandler: (response) => response.blob(),
      }),
    }),
  }),
});

export const {
  useGetDashboardSummaryQuery,
  useGetMetricCardsQuery,
  useGetMetricTrendQuery,
  useGetRegionsQuery,
  useGetRegionQuery,
  useGetHeatmapDataQuery,
  useGetAnalyticsQuery,
  useGetComparisonQuery,
  useExportDashboardMutation,
} = dashboardApi;
