import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { Simulation, SimulationResult } from '../../types';

export const simulationApi = createApi({
  reducerPath: 'simulationApi',
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
    getSimulations: builder.query<Simulation[], { status?: string; limit?: number }>({
      query: (params) => {
        const queryParams = new URLSearchParams();
        if (params.status) queryParams.append('status', params.status);
        if (params.limit) queryParams.append('limit', params.limit.toString());
        return `/simulations?${queryParams.toString()}`;
      },
    }),
    getSimulation: builder.query<Simulation, string>({
      query: (simulationId) => `/simulations/${simulationId}`,
    }),
    createSimulation: builder.mutation<Simulation, Partial<Simulation>>({
      query: (simulation) => ({
        url: '/simulations',
        method: 'POST',
        body: simulation,
      }),
    }),
    runSimulation: builder.mutation<SimulationResult, string>({
      query: (simulationId) => ({
        url: `/simulations/${simulationId}/run`,
        method: 'POST',
      }),
    }),
    getSimulationResult: builder.query<SimulationResult, string>({
      query: (simulationId) => `/simulations/${simulationId}/result`,
    }),
    deleteSimulation: builder.mutation<void, string>({
      query: (simulationId) => ({
        url: `/simulations/${simulationId}`,
        method: 'DELETE',
      }),
    }),
    exportSimulation: builder.mutation<Blob, { id: string; format: string }>({
      query: ({ id, format }) => ({
        url: `/simulations/${id}/export`,
        method: 'POST',
        body: { format },
        responseHandler: (response) => response.blob(),
      }),
    }),
  }),
});

export const {
  useGetSimulationsQuery,
  useGetSimulationQuery,
  useCreateSimulationMutation,
  useRunSimulationMutation,
  useGetSimulationResultQuery,
  useDeleteSimulationMutation,
  useExportSimulationMutation,
} = simulationApi;
