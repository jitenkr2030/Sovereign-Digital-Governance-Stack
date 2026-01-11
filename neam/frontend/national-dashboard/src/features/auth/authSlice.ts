import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import { authApi } from '../api/authApi';
import type { User, AuthState, LoginCredentials, JwtPayload } from './types';
import type { RootState } from '../../store';

const initialState: AuthState = {
  user: null,
  token: null,
  refreshToken: null,
  isAuthenticated: false,
  isLoading: false,
  error: null,
  sessionExpiry: null,
};

// Async thunk for user login
export const login = createAsyncThunk<
  { user: User; token: string; refreshToken: string },
  LoginCredentials,
  { rejectValue: string }
>(
  'auth/login',
  async (credentials, { rejectWithValue }) => {
    try {
      const response = await authApi.login(credentials);
      return response;
    } catch (error) {
      if (error instanceof Error) {
        return rejectWithValue(error.message);
      }
      return rejectWithValue('An unknown error occurred during login');
    }
  }
);

// Async thunk for refreshing the access token
export const refreshAccessToken = createAsyncThunk<
  { token: string },
  void,
  { rejectValue: string; state: { auth: AuthState } }
>(
  'auth/refreshToken',
  async (_, { getState, rejectWithValue }) => {
    const { auth } = getState();
    if (!auth.refreshToken) {
      return rejectWithValue('No refresh token available');
    }

    try {
      const response = await authApi.refreshToken(auth.refreshToken);
      return response;
    } catch (error) {
      if (error instanceof Error) {
        return rejectWithValue(error.message);
      }
      return rejectWithValue('Failed to refresh token');
    }
  }
);

// Async thunk for fetching current user profile
export const fetchCurrentUser = createAsyncThunk<
  User,
  void,
  { rejectValue: string; state: { auth: AuthState } }
>(
  'auth/fetchCurrentUser',
  async (_, { getState, rejectWithValue }) => {
    const { auth } = getState();
    if (!auth.token) {
      return rejectWithValue('No authentication token available');
    }

    try {
      const user = await authApi.getCurrentUser();
      return user;
    } catch (error) {
      if (error instanceof Error) {
        return rejectWithValue(error.message);
      }
      return rejectWithValue('Failed to fetch user profile');
    }
  }
);

// Helper function to decode JWT and get expiry time
function getTokenExpiry(token: string): Date | null {
  try {
    const payloadBase64 = token.split('.')[1];
    if (!payloadBase64) return null;

    const payload = JSON.parse(atob(payloadBase64));
    if (payload.exp) {
      return new Date(payload.exp * 1000);
    }
    return null;
  } catch {
    return null;
  }
}

// Helper function to check if token is expired
function isTokenExpired(expiry: Date | null): boolean {
  if (!expiry) return true;
  // Add a buffer of 5 minutes before actual expiry
  const buffer = 5 * 60 * 1000;
  return new Date().getTime() + buffer >= expiry.getTime();
}

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    // Clear authentication state on logout
    logout: (state) => {
      state.user = null;
      state.token = null;
      state.refreshToken = null;
      state.isAuthenticated = false;
      state.error = null;
      state.sessionExpiry = null;
    },
    // Clear any authentication errors
    clearError: (state) => {
      state.error = null;
    },
    // Update user profile in state (after profile update)
    updateUserProfile: (state, action: PayloadAction<Partial<User>>) => {
      if (state.user) {
        state.user = { ...state.user, ...action.payload };
      }
    },
    // Set authentication state from storage (for session persistence)
    restoreSession: (
      state,
      action: PayloadAction<{ user: User; token: string; refreshToken: string }>
    ) => {
      const { user, token, refreshToken } = action.payload;
      state.user = user;
      state.token = token;
      state.refreshToken = refreshToken;
      state.isAuthenticated = true;
      state.sessionExpiry = getTokenExpiry(token);
    },
  },
  extraReducers: (builder) => {
    // Login lifecycle
    builder
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload.user;
        state.token = action.payload.token;
        state.refreshToken = action.payload.refreshToken;
        state.isAuthenticated = true;
        state.sessionExpiry = getTokenExpiry(action.payload.token);
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload || 'Login failed';
        state.isAuthenticated = false;
      });

    // Token refresh lifecycle
    builder
      .addCase(refreshAccessToken.fulfilled, (state, action) => {
        state.token = action.payload.token;
        state.sessionExpiry = getTokenExpiry(action.payload.token);
      })
      .addCase(refreshAccessToken.rejected, (state) => {
        // If token refresh fails, log the user out
        state.user = null;
        state.token = null;
        state.refreshToken = null;
        state.isAuthenticated = false;
        state.sessionExpiry = null;
      });

    // Fetch current user lifecycle
    builder
      .addCase(fetchCurrentUser.pending, (state) => {
        state.isLoading = true;
      })
      .addCase(fetchCurrentUser.fulfilled, (state, action) => {
        state.isLoading = false;
        state.user = action.payload;
      })
      .addCase(fetchCurrentUser.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload || 'Failed to fetch user';
      });
  },
});

// Export actions
export const { logout, clearError, updateUserProfile, restoreSession } =
  authSlice.actions;

// Selectors
export const selectCurrentUser = (state: RootState): User | null =>
  state.auth.user;

export const selectIsAuthenticated = (state: RootState): boolean =>
  state.auth.isAuthenticated;

export const selectAuthLoading = (state: RootState): boolean =>
  state.auth.isLoading;

export const selectAuthError = (state: RootState): string | null =>
  state.auth.error;

export const selectAuthToken = (state: RootState): string | null =>
  state.auth.token;

export const selectSessionExpiry = (state: RootState): Date | null =>
  state.auth.sessionExpiry;

export const selectIsSessionExpired = (state: RootState): boolean =>
  isTokenExpired(state.auth.sessionExpiry);

// Export reducer
export default authSlice.reducer;
