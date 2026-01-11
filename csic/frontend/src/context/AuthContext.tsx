import React, { createContext, useContext, useReducer, useEffect, useCallback } from 'react';
import type { User, LoginCredentials, AuthState, UserRole } from '../types';

type AuthAction =
  | { type: 'LOGIN_START' }
  | { type: 'LOGIN_SUCCESS'; payload: { user: User; token: string } }
  | { type: 'LOGIN_FAILURE'; payload: string }
  | { type: 'LOGOUT' }
  | { type: 'SET_USER'; payload: User }
  | { type: 'TOKEN_REFRESHED'; payload: string };

interface AuthContextType extends AuthState {
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
  hasRole: (roles: UserRole | UserRole[]) => boolean;
}

const initialState: AuthState = {
  user: null,
  isAuthenticated: false,
  isLoading: true,
  token: null,
};

function authReducer(state: AuthState, action: AuthAction): AuthState {
  switch (action.type) {
    case 'LOGIN_START':
      return { ...state, isLoading: true };
    case 'LOGIN_SUCCESS':
      return {
        ...state,
        isLoading: false,
        isAuthenticated: true,
        user: action.payload.user,
        token: action.payload.token,
      };
    case 'LOGIN_FAILURE':
      return {
        ...state,
        isLoading: false,
        isAuthenticated: false,
        user: null,
        token: null,
      };
    case 'LOGOUT':
      return { ...initialState, isLoading: false };
    case 'SET_USER':
      return { ...state, user: action.payload };
    case 'TOKEN_REFRESHED':
      return { ...state, token: action.payload };
    default:
      return state;
  }
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Permission definitions for each role
const rolePermissions: Record<UserRole, string[]> = {
  admin: [
    'users:read', 'users:write', 'users:delete',
    'licenses:read', 'licenses:write', 'licenses:revoke',
    'keys:read', 'keys:write', 'keys:rotate', 'keys:delete',
    'audit:read', 'audit:export',
    'reports:read', 'reports:write',
    'siem:read', 'siem:configure',
    'compliance:read', 'compliance:manage',
    'settings:read', 'settings:write',
    'auditor:verify', 'auditor:chain',
  ],
  regulator: [
    'licenses:read',
    'audit:read', 'audit:export',
    'reports:read', 'reports:generate',
    'compliance:read',
    'siem:read',
    'auditor:verify', 'auditor:chain',
  ],
  auditor: [
    'audit:read', 'audit:export',
    'keys:read',
    'reports:read', 'reports:generate',
    'auditor:verify', 'auditor:chain',
  ],
  user: [
    'licenses:read',
    'audit:read:own',
    'keys:read:own',
  ],
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Check for existing session on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const token = localStorage.getItem('auth_token');
        const userStr = localStorage.getItem('user_data');

        if (token && userStr) {
          const user = JSON.parse(userStr);
          dispatch({ type: 'SET_USER', payload: user });
        }
      } catch {
        localStorage.removeItem('auth_token');
        localStorage.removeItem('user_data');
      } finally {
        dispatch({ type: 'LOGIN_SUCCESS', payload: { user: state.user!, token: state.token! } });
      }
    };

    checkAuth();
  }, []);

  const login = useCallback(async (credentials: LoginCredentials) => {
    dispatch({ type: 'LOGIN_START' });

    try {
      // Simulate API call - replace with actual API
      await new Promise(resolve => setTimeout(resolve, 1000));

      // Mock successful login
      const mockUser: User = {
        id: '1',
        email: credentials.email,
        name: 'Demo User',
        role: credentials.email.includes('admin') ? 'admin' : 
              credentials.email.includes('regulator') ? 'regulator' :
              credentials.email.includes('auditor') ? 'auditor' : 'user',
        organization: 'CSIC Corp',
        permissions: [],
        lastLogin: new Date().toISOString(),
        mfaEnabled: false,
      };

      const mockToken = 'mock_jwt_token_' + Date.now();

      localStorage.setItem('auth_token', mockToken);
      localStorage.setItem('user_data', JSON.stringify(mockUser));

      dispatch({ type: 'LOGIN_SUCCESS', payload: { user: mockUser, token: mockToken } });
    } catch (error) {
      dispatch({ type: 'LOGIN_FAILURE', payload: 'Login failed' });
      throw error;
    }
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user_data');
    dispatch({ type: 'LOGOUT' });
  }, []);

  const hasPermission = useCallback((permission: string): boolean => {
    if (!state.user) return false;
    const userPermissions = rolePermissions[state.user.role] || [];
    return userPermissions.includes(permission) || 
           userPermissions.some(p => permission.startsWith(p.split(':')[0] + ':'));
  }, [state.user]);

  const hasRole = useCallback((roles: UserRole | UserRole[]): boolean => {
    if (!state.user) return false;
    const roleArray = Array.isArray(roles) ? roles : [roles];
    return roleArray.includes(state.user.role);
  }, [state.user]);

  return (
    <AuthContext.Provider value={{ ...state, login, logout, hasPermission, hasRole }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
