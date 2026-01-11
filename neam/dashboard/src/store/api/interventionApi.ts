import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Intervention, InterventionStatus } from '../../types';

export const interventionApi = createApi({
  reducerPath: 'interventionApi',
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
    getInterventions: builder.query<Intervention[], { status?: InterventionStatus; regionId?: string; limit?: number }>({
      query: (params) => {
        const queryParams = new URLSearchParams();
        if (params.status) queryParams.append('status', params.status);
        if (params.regionId) queryParams.append('regionId', params.regionId);
        if (params.limit) queryParams.append('limit', params.limit.toString());
        return `/interventions?${queryParams.toString()}`;
      },
    }),
    getIntervention: builder.query<Intervention, string>({
      query: (interventionId) => `/interventions/${interventionId}`,
    }),
    getActiveInterventions: builder.query<Intervention[], void>({
      query: () => '/interventions/active',
    }),
    approveIntervention: builder.mutation<Intervention, { id: string; approver: string; notes?: string }>({
      query: ({ id, approver, notes }) => ({
        url: `/interventions/${id}/approve`,
        method: 'POST',
        body: { approver, notes },
      }),
    }),
    cancelIntervention: builder.mutation<Intervention, { id: string; reason: string }>({
      query: ({ id, reason }) => ({
        url: `/interventions/${id}/cancel`,
        method: 'POST',
        body: { reason },
      }),
    }),
    triggerIntervention: builder.mutation<Intervention, { policyId: string; regionId: string; manualTrigger: boolean }>({
      query: (params) => ({
        url: '/interventions/trigger',
        method: 'POST',
        body: params,
      }),
    }),
    getInterventionHistory: builder.query<Intervention[], string>({
      query: (interventionId) => `/interventions/${interventionId}/history`,
    }),
    getInterventionStats: builder.query<{
      total: number;
      byStatus: Record<string, number>;
      byPolicy: Record<string, number>;
      avgCompletionTime: number;
    }, void>({
      query: () => '/interventions/stats',
    }),
  }),
});

export const {
  useGetInterventionsQuery,
  useGetInterventionQuery,
  useGetActiveInterventionsQuery,
  useApproveInterventionMutation,
  useCancelInterventionMutation,
  useTriggerInterventionMutation,
  useGetInterventionHistoryQuery,
  useGetInterventionStatsQuery,
} = interventionApi;
