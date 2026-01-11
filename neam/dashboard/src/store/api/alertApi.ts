import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Alert, AlertType } from '../../types';

export const alertApi = createApi({
  reducerPath: 'alertApi',
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
    getAlerts: builder.query<Alert[], { types?: AlertType[]; status?: string; limit?: number }>({
      query: (params) => {
        const queryParams = new URLSearchParams();
        if (params.types) queryParams.append('types', params.types.join(','));
        if (params.status) queryParams.append('status', params.status);
        if (params.limit) queryParams.append('limit', params.limit.toString());
        return `/alerts?${queryParams.toString()}`;
      },
    }),
    getAlert: builder.query<Alert, string>({
      query: (alertId) => `/alerts/${alertId}`,
    }),
    acknowledgeAlert: builder.mutation<Alert, { id: string; notes?: string }>({
      query: ({ id, notes }) => ({
        url: `/alerts/${id}/acknowledge`,
        method: 'POST',
        body: { notes },
      }),
    }),
    resolveAlert: builder.mutation<Alert, { id: string; resolution: string }>({
      query: ({ id, resolution }) => ({
        url: `/alerts/${id}/resolve`,
        method: 'POST',
        body: { resolution },
      }),
    }),
    getAlertStats: builder.query<{ total: number; byType: Record<string, number>; byPriority: Record<string, number> }, void>({
      query: () => '/alerts/stats',
    }),
    subscribeToAlerts: builder.query<Alert[], void>({
      query: () => '/alerts/subscribed',
    }),
  }),
});

export const {
  useGetAlertsQuery,
  useGetAlertQuery,
  useAcknowledgeAlertMutation,
  useResolveAlertMutation,
  useGetAlertStatsQuery,
  useSubscribeToAlertsQuery,
} = alertApi;
