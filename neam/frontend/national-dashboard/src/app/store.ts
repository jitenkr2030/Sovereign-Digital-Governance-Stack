/**
 * Redux Store Configuration
 * 
 * Configures the Redux store with middleware, devtools, and type definitions
 * for the NEAM frontend application.
 */

import { configureStore, combineReducers, Middleware } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { apiSlice } from '../features/api/apiSlice';
import authReducer from '../features/auth/authSlice';
import uiReducer from '../features/ui/uiSlice';

/**
 * Combine all reducers into a root reducer
 */
const rootReducer = combineReducers({
  [apiSlice.reducerPath]: apiSlice.reducer,
  auth: authReducer,
  ui: uiReducer,
});

/**
 * Custom middleware to log Redux actions in development
 */
const actionLogger: Middleware = (store) => (next) => (action) => {
  if (process.env.NODE_ENV === 'development') {
    console.log('Dispatching action:', action);
  }
  return next(action);
};

/**
 * Custom error handling middleware for API errors
 */
const errorHandler: Middleware = (store) => (next) => (action) => {
  // Handle API-related errors globally
  if (action.type?.startsWith('api/') && action.type.includes('rejected')) {
    const error = action.payload as { status: number; data?: { message?: string } };
    
    if (error.status === 401) {
      // Unauthorized - trigger logout
      store.dispatch({ type: 'auth/logout' });
    }
    
    // Log error for monitoring
    console.error('API Error:', error);
  }
  
  return next(action);
};

/**
 * Configure and create the Redux store
 */
export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types for serialization check
        ignoredActions: ['auth/setCredentials', 'persist/PERSIST'],
        // Ignore these paths in the state
        ignoredPaths: ['auth.token', 'auth.refreshToken'],
      },
      immutableCheck: {
        // Ignore certain paths for performance
        ignoredPaths: ['api.queries', 'api.mutations'],
      },
    }).concat(apiSlice.middleware, actionLogger, errorHandler),
  devTools: process.env.NODE_ENV !== 'production',
});

// Configure listeners for RTK Query cache behavior
setupListeners(store.dispatch);

/**
 * Type definitions for TypeScript
 */
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
export type AppStore = typeof store;

/**
 * Preloaded state for hydration (e.g., from localStorage)
 */
export interface PreloadedState {
  auth: ReturnType<typeof authReducer>;
  ui: ReturnType<typeof uiReducer>;
}

/**
 * Create a store instance with preloaded state
 * Used for testing and SSR
 */
export function createStore(preloadedState?: PreloadedState) {
  return configureStore({
    reducer: rootReducer,
    middleware: (getDefaultMiddleware) =>
      getDefaultMiddleware({
        serializableCheck: {
          ignoredActions: ['auth/setCredentials', 'persist/PERSIST'],
          ignoredPaths: ['auth.token', 'auth.refreshToken'],
        },
      }).concat(apiSlice.middleware, actionLogger, errorHandler),
    devTools: process.env.NODE_ENV !== 'production',
    preloadedState,
  });
}

export default store;
