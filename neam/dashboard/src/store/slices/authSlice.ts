import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import type { User, UserRole, Permission } from '../../types';
import { authApi } from '../api/authApi';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
}

const initialState: AuthState = {
  user: null,
  token: localStorage.getItem('neam_token'),
  isAuthenticated: !!localStorage.getItem('neam_token'),
  isLoading: false,
  error: null,
};

export const login = createAsyncThunk(
  'auth/login',
  async (credentials: { email: string; password: string }, { rejectWithValue }) => {
    try {
      const response = await authApi.login(credentials.email, credentials.password);
      localStorage.setItem('neam_token', response.token);
      return response;
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  }
);

export const logout = createAsyncThunk('auth/logout', async () => {
  localStorage.removeItem('neam_token');
});

export const refreshUser = createAsyncThunk(
  'auth/refreshUser',
  async (_, { rejectWithValue }) => {
    try {
      const response = await authApi.getCurrentUser();
      return response;
    } catch (error) {
      return rejectWithValue((error as Error).message);
    }
  }
);

const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
    setToken: (state, action: PayloadAction<string>) => {
      state.token = action.payload;
      state.isAuthenticated = true;
      localStorage.setItem('neam_token', action.payload);
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(login.pending, (state) => {
        state.isLoading = true;
        state.error = null;
      })
      .addCase(login.fulfilled, (state, action) => {
        state.isLoading = false;
        state.isAuthenticated = true;
        state.user = action.payload.user;
        state.token = action.payload.token;
      })
      .addCase(login.rejected, (state, action) => {
        state.isLoading = false;
        state.error = action.payload as string;
      })
      .addCase(logout.fulfilled, (state) => {
        state.user = null;
        state.token = null;
        state.isAuthenticated = false;
      })
      .addCase(refreshUser.fulfilled, (state, action) => {
        state.user = action.payload;
      });
  },
});

export const { clearError, setToken } = authSlice.actions;
export default authSlice.reducer;

// Role-based access control helpers
export const hasPermission = (user: User | null, permission: Permission): boolean => {
  if (!user) return false;
  return user.permissions.includes(permission);
};

export const hasRole = (user: User | null, roles: UserRole[]): boolean => {
  if (!user) return false;
  return roles.includes(user.role);
};

export const getRoleDisplayName = (role: UserRole): string => {
  const roleNames: Record<UserRole, string> = {
    PMO: 'Prime Minister\'s Office',
    FINANCE: 'Ministry of Finance',
    STATE_ADMIN: 'State Administrator',
    DISTRICT_COLLECTOR: 'District Collector',
    ANALYST: 'Economic Analyst',
  };
  return roleNames[role];
};

export const getRolePermissions = (role: UserRole): Permission[] => {
  const permissionsByRole: Record<UserRole, Permission[]> = {
    PMO: [
      'VIEW_DASHBOARD', 'VIEW_NATIONAL', 'VIEW_REGIONAL', 'VIEW_DISTRICT',
      'CREATE_POLICY', 'EDIT_POLICY', 'DELETE_POLICY', 'APPROVE_INTERVENTION',
      'TRIGGER_INTERVENTION', 'CANCEL_INTERVENTION', 'CREATE_SIMULATION',
      'VIEW_ANALYTICS', 'EXPORT_DATA', 'MANAGE_USERS',
    ],
    FINANCE: [
      'VIEW_DASHBOARD', 'VIEW_NATIONAL', 'VIEW_REGIONAL',
      'CREATE_POLICY', 'EDIT_POLICY', 'APPROVE_INTERVENTION',
      'CREATE_SIMULATION', 'VIEW_ANALYTICS', 'EXPORT_DATA',
    ],
    STATE_ADMIN: [
      'VIEW_DASHBOARD', 'VIEW_REGIONAL', 'VIEW_DISTRICT',
      'EDIT_POLICY', 'APPROVE_INTERVENTION', 'TRIGGER_INTERVENTION',
      'VIEW_ANALYTICS', 'EXPORT_DATA',
    ],
    DISTRICT_COLLECTOR: [
      'VIEW_DASHBOARD', 'VIEW_DISTRICT', 'VIEW_ANALYTICS',
    ],
    ANALYST: [
      'VIEW_DASHBOARD', 'VIEW_NATIONAL', 'VIEW_REGIONAL',
      'CREATE_SIMULATION', 'VIEW_ANALYTICS', 'EXPORT_DATA',
    ],
  };
  return permissionsByRole[role] || [];
};
