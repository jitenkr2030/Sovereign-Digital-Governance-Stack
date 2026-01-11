import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Policy, PolicyCategory, PolicyStatus } from '../../types';

export const policyApi = createApi({
  reducerPath: 'policyApi',
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
    getPolicies: builder.query<Policy[], { category?: PolicyCategory; status?: PolicyStatus }>({
      query: (params) => {
        const queryParams = new URLSearchParams();
        if (params.category) queryParams.append('category', params.category);
        if (params.status) queryParams.append('status', params.status);
        return `/policies?${queryParams.toString()}`;
      },
    }),
    getPolicy: builder.query<Policy, string>({
      query: (policyId) => `/policies/${policyId}`,
    }),
    createPolicy: builder.mutation<Policy, Partial<Policy>>({
      query: (policy) => ({
        url: '/policies',
        method: 'POST',
        body: policy,
      }),
    }),
    updatePolicy: builder.mutation<Policy, { id: string; policy: Partial<Policy> }>({
      query: ({ id, policy }) => ({
        url: `/policies/${id}`,
        method: 'PUT',
        body: policy,
      }),
    }),
    deletePolicy: builder.mutation<void, string>({
      query: (policyId) => ({
        url: `/policies/${policyId}`,
        method: 'DELETE',
      }),
    }),
    activatePolicy: builder.mutation<void, string>({
      query: (policyId) => ({
        url: `/policies/${policyId}/activate`,
        method: 'POST',
      }),
    }),
    deactivatePolicy: builder.mutation<void, string>({
      query: (policyId) => ({
        url: `/policies/${policyId}/deactivate`,
        method: 'POST',
      }),
    }),
    getPolicyHistory: builder.query<Policy[], string>({
      query: (policyId) => `/policies/${policyId}/history`,
    }),
    validatePolicy: builder.mutation<{ valid: boolean; errors: string[] }, Partial<Policy>>({
      query: (policy) => ({
        url: '/policies/validate',
        method: 'POST',
        body: policy,
      }),
    }),
  }),
});

export const {
  useGetPoliciesQuery,
  useGetPolicyQuery,
  useCreatePolicyMutation,
  useUpdatePolicyMutation,
  useDeletePolicyMutation,
  useActivatePolicyMutation,
  useDeactivatePolicyMutation,
  useGetPolicyHistoryQuery,
  useValidatePolicyMutation,
} = policyApi;
