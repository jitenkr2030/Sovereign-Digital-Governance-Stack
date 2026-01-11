import { configureStore } from '@reduxjs/toolkit';
import { setupListeners } from '@reduxjs/toolkit/query';
import { authApi } from './api/authApi';
import { dashboardApi } from './api/dashboardApi';
import { policyApi } from './api/policyApi';
import { alertApi } from './api/alertApi';
import { interventionApi } from './api/interventionApi';
import { simulationApi } from './api/simulationApi';
import authReducer from './slices/authSlice';
import dashboardReducer from './slices/dashboardSlice';
import alertReducer from './slices/alertSlice';
import notificationReducer from './slices/notificationSlice';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    dashboard: dashboardReducer,
    alerts: alertReducer,
    notifications: notificationReducer,
    [authApi.reducerPath]: authApi.reducer,
    [dashboardApi.reducerPath]: dashboardApi.reducer,
    [policyApi.reducerPath]: policyApi.reducer,
    [alertApi.reducerPath]: alertApi.reducer,
    [interventionApi.reducerPath]: interventionApi.reducer,
    [simulationApi.reducerPath]: simulationApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(authApi.middleware)
      .concat(dashboardApi.middleware)
      .concat(policyApi.middleware)
      .concat(alertApi.middleware)
      .concat(interventionApi.middleware)
      .concat(simulationApi.middleware),
  devTools: import.meta.env.DEV,
});

setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
