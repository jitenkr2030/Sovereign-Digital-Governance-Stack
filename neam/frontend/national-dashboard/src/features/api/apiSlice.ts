import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';
import type { RootState } from '../../store';
import { logout, selectAuthToken } from '../auth/authSlice';
import type { User, LoginCredentials, TokenResponse } from './types';

// Base URL for API requests - configurable via environment variables
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || '/api/v1';

// Tag types for cache invalidation
export enum CacheTags {
  USERS = 'Users',
  DASHBOARDS = 'Dashboards',
  REPORTS = 'Reports',
  NOTIFICATIONS = 'Notifications',
  SETTINGS = 'Settings',
}

// Base query with authentication and error handling
const baseQuery = fetchBaseQuery({
  baseUrl: API_BASE_URL,
  prepareHeaders: (headers, { getState }) => {
    const token = selectAuthToken(getState() as RootState);
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    }
    headers.set('Content-Type', 'application/json');
    return headers;
  },
});

// Enhanced base query with token refresh capability
const baseQueryWithReauth = async (
  args: string | Request,
  api: { getState: () => unknown; dispatch: (action: unknown) => unknown },
  extraOptions: Record<string, unknown>
) => {
  let result = await baseQuery(args, api as Parameters<typeof baseQuery>[1], extraOptions);

  // If we get a 401 error, try to refresh the token
  if (result.error && result.error.status === 401) {
    const dispatch = api.dispatch as (action: { payload?: unknown; type: string }) => void;
    const getState = api.getState as () => { auth: { refreshToken: string | null } };

    const refreshToken = getState().auth.refreshToken;

    if (refreshToken) {
      try {
        // Attempt to refresh the token
        const refreshResponse = await baseQuery(
          {
            url: '/auth/refresh',
            method: 'POST',
            body: { refreshToken },
          },
          api,
          extraOptions
        );

        if (refreshResponse.data) {
          const { token } = refreshResponse.data as TokenResponse;
          // Update the stored token (this is a simplified approach)
          // In production, you'd dispatch an action to update the token in state
          localStorage.setItem('accessToken', token);

          // Retry the original request with the new token
          result = await baseQuery(
            args,
            {
              ...api,
              getState: () => ({ ...getState(), auth: { ...getState().auth, token } }),
            },
            extraOptions
          );
        } else {
          // Refresh failed, logout the user
          dispatch(logout());
        }
      } catch (refreshError) {
        dispatch(logout());
      }
    } else {
      // No refresh token available, logout
      dispatch(logout());
    }
  }

  return result;
};

// Create the main API slice
export const apiSlice = createApi({
  reducerPath: 'api',
  baseQuery: baseQueryWithReauth,
  tagTypes: Object.values(CacheTags),
  endpoints: (builder) => ({}),
});

// Auth API endpoints
export const authApi = createApi({
  reducerPath: 'authApi',
  baseQuery: baseQueryWithReauth,
  tagTypes: [CacheTags.USERS],
  endpoints: (builder) => ({
    // Login endpoint
    login: builder.mutation<TokenResponse & { user: User }, LoginCredentials>({
      query: (credentials) => ({
        url: '/auth/login',
        method: 'POST',
        body: credentials,
      }),
    }),
    // Logout endpoint
    logout: builder.mutation<void, void>({
      query: () => ({
        url: '/auth/logout',
        method: 'POST',
      }),
    }),
    // Refresh token endpoint
    refreshToken: builder.mutation<TokenResponse, string>({
      query: (refreshToken) => ({
        url: '/auth/refresh',
        method: 'POST',
        body: { refreshToken },
      }),
    }),
    // Get current user profile
    getCurrentUser: builder.query<User, void>({
      query: () => '/auth/me',
      providesTags: [{ type: CacheTags.USERS, id: 'current' }],
    }),
    // Update current user profile
    updateProfile: builder.mutation<User, Partial<User>>({
      query: (profileData) => ({
        url: '/auth/profile',
        method: 'PUT',
        body: profileData,
      }),
      invalidatesTags: [{ type: CacheTags.USERS, id: 'current' }],
    }),
    // Change password
    changePassword: builder.mutation<
      void,
      { currentPassword: string; newPassword: string }
    >({
      query: (passwordData) => ({
        url: '/auth/password',
        method: 'PUT',
        body: passwordData,
      }),
    }),
  }),
});

// Export hooks for usage in components
export const {
  useLoginMutation,
  useLogoutMutation,
  useRefreshTokenMutation,
  useGetCurrentUserQuery,
  useUpdateProfileMutation,
  useChangePasswordMutation,
} = authApi;

// Export the api slice for injection
export { baseQueryWithReauth, API_BASE_URL };
