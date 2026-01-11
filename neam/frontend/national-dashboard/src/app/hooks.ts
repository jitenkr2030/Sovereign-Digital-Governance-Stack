/**
 * Typed Redux Hooks
 * 
 * Provides TypeScript-typed versions of useDispatch and useSelector
 * for the Redux store. This ensures type safety throughout the app.
 */

import { TypedUseSelectorHook, useDispatch, useSelector } from 'react-redux';
import type { RootState, AppDispatch } from './store';

/**
 * Typed useDispatch hook
 * 
 * Use this instead of the standard useDispatch hook to get
 * proper TypeScript type inference for dispatch actions.
 * 
 * @example
 * const dispatch = useAppDispatch();
 * dispatch(authActions.setCredentials({ token, user }));
 */
export const useAppDispatch: () => AppDispatch = useDispatch;

/**
 * Typed useSelector hook
 * 
 * Use this instead of the standard useSelector hook to get
 * proper TypeScript type inference for selector results.
 * 
 * @example
 * const user = useAppSelector((state) => state.auth.user);
 * const isAuthenticated = useAppSelector(selectIsAuthenticated);
 */
export const useAppSelector: TypedUseSelectorHook<RootState> = useSelector;

/**
 * Selector hook factories for common selections
 * These can be used to avoid repeated selector logic
 */

/**
 * Select the current user from auth state
 */
export const selectCurrentUser = (state: RootState) => state.auth.user;

/**
 * Select authentication token
 */
export const selectAuthToken = (state: RootState) => state.auth.token;

/**
 * Select authentication status
 */
export const selectIsAuthenticated = (state: RootState) => 
  state.auth.isAuthenticated && !state.auth.loading;

/**
 * Select user roles
 */
export const selectUserRoles = (state: RootState) => 
  state.auth.user?.roles || [];

/**
 * Select auth loading state
 */
export const selectAuthLoading = (state: RootState) => 
  state.auth.loading === 'pending';

/**
 * Select auth error
 */
export const selectAuthError = (state: RootState) => state.auth.error;

/**
 * Select UI theme
 */
export const selectTheme = (state: RootState) => state.ui.theme;

/**
 * Select sidebar state
 */
export const selectSidebarCollapsed = (state: RootState) => 
  state.ui.sidebarCollapsed;

/**
 * Select active modal
 */
export const selectActiveModal = (state: RootState) => 
  state.ui.activeModal;

/**
 * Select notifications
 */
export const selectNotifications = (state: RootState) => 
  state.ui.notifications;

/**
 * Select API cache state
 */
export const selectApiState = (state: RootState) => state.api;

/**
 * Check if user has a specific role
 */
export const selectHasRole = (role: string) => (state: RootState) => 
  state.auth.user?.roles.includes(role) || false;

/**
 * Check if user has any of the specified roles
 */
export const selectHasAnyRole = (roles: string[]) => (state: RootState) => 
  state.auth.user?.roles.some((r) => roles.includes(r)) || false;

/**
 * Check if user has all specified roles
 */
export const selectHasAllRoles = (roles: string[]) => (state: RootState) => {
  const userRoles = state.auth.user?.roles || [];
  return roles.every((r) => userRoles.includes(r));
};

/**
 * Select session expiration status
 */
export const selectSessionExpired = (state: RootState) => {
  if (!state.auth.sessionExpiresAt) return false;
  return Date.now() > state.auth.sessionExpiresAt;
};

/**
 * Get remaining session time in milliseconds
 */
export const selectSessionRemaining = (state: RootState) => {
  if (!state.auth.sessionExpiresAt) return 0;
  return Math.max(0, state.auth.sessionExpiresAt - Date.now());
};
