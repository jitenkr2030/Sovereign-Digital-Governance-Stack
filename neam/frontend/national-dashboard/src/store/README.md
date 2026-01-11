/**
 * NEAM Frontend State Management Architecture
 * 
 * This document describes the Redux Toolkit-based state management architecture
 * for the National Economic Activity Monitor frontend application.
 * 
 * ## Architecture Overview
 * 
 * The application uses Redux Toolkit (RTK) as its primary state management solution,
 * complementing React Query for server state management. This hybrid approach
 * provides:
 * - Centralized client state (user session, UI preferences, notifications)
 * - Efficient server state caching with automatic refetching (RTK Query)
 * - Predictable state updates with Immer-based mutable logic
 * - Time-travel debugging with Redux DevTools
 * 
 * ## Directory Structure
 * 
 * ```
 * src/
 * ├── app/
 * │   ├── store.ts              # Redux store configuration
 * │   ├── hooks.ts              # Typed dispatch/selector hooks
 * │   └── rootReducer.ts        # Combined reducer interface
 * ├── features/
 * │   ├── auth/
 * │   │   ├── authSlice.ts      # Authentication state & actions
 * │   │   ├── authApi.ts        # Auth-related API endpoints
 * │   │   ├── types.ts          # Auth type definitions
 * │   │   └── utils.ts          # JWT & token utilities
 * │   ├── api/
 * │   │   ├── apiSlice.ts       # Base RTK Query configuration
 * │   │   └── endpoints.ts      # Feature API endpoints
 * │   └── ui/
 * │       ├── uiSlice.ts        # UI state (theme, sidebar, modals)
 * │       └── notifications.ts  # Notification system
 * ├── guards/
 * │   ├── AuthGuard.tsx         # Route guard for authenticated users
 * │   ├── RoleGuard.tsx         # Route guard for role-based access
 * │   ├── GuestGuard.tsx        # Route guard for public routes
 * │   └── index.ts              # Guard exports
 * ├── routes/
 * │   ├── AppRoutes.tsx         # Main routing configuration
 * │   ├── ProtectedRoute.tsx    # Protected route wrapper
 * │   └── RouteConstants.ts     # Route path constants
 * └── utils/
 *     ├── errorBoundary.tsx     # Error boundary with Redux integration
 *     └── localStorage.ts       # Persistence utilities
 * ```
 * 
 * ## State Categories
 * 
 * ### 1. Authentication State (authSlice)
 * 
 * The authentication slice manages the user's session state including:
 * - User profile information (from JWT claims)
 * - Access and refresh tokens
 * - Authentication status (loading, authenticated, error)
 * - Session expiration tracking
 * 
 * ```typescript
 * interface AuthState {
 *   user: User | null;
 *   token: string | null;
 *   refreshToken: string | null;
 *   isAuthenticated: boolean;
 *   loading: 'idle' | 'pending' | 'succeeded' | 'failed';
 *   error: string | null;
 *   sessionExpiresAt: number | null;
 * }
 * ```
 * 
 * ### 2. UI State (uiSlice)
 * 
 * UI state manages application-level preferences:
 * - Theme mode (light/dark/system)
 * - Sidebar state (collapsed/expanded)
 * - Active modals and dialogs
 * - Global loading indicators
 * 
 * ### 3. Server State (RTK Query)
 * 
 * RTK Query manages server state with automatic caching:
 * - Automatic polling for real-time updates
 * - Optimistic updates for mutations
 * - Cache invalidation based on tags
 * - Parallel and dependent query handling
 * 
 * ## Integration Points
 * 
 * ### Error Boundaries
 * 
 * Error boundaries integrate with Redux to log and track UI errors:
 * - Component errors are dispatched to Redux
 * - Error state is displayed via notification system
 * - Recovery mechanisms allow retrying failed operations
 * 
 * ### Persistence
 * 
 * Authentication state persists across page refreshes:
 * - Tokens stored in localStorage (encrypted)
 * - User state rehydrated on app initialization
 * - Session expiration checked on app load
 * 
 * ## Best Practices
 * 
 * 1. **Keep State Normalized**: Use entity adapters for collections
 * 2. **Minimize Derived State**: Use selectors for computed values
 * 3. **Use Immer**: Mutate state directly with Immer proxies
 * 4. **Handle Async**: Use createAsyncThunk for side effects
 * 5. **Type Everything**: Define strict TypeScript interfaces
 * 6. **Document Slices**: Add JSDoc comments for each slice
 * 
 * ## Performance Considerations
 * 
 * - Selectors memoize derived data automatically
 * - RTK Query prevents duplicate requests
 * - Entity adapters optimize collection lookups
 * - Middleware runs only on relevant state changes
 * 
 * ## Security Notes
 * 
 * - Sensitive data in Redux DevTools is masked in production
 * - Token refresh happens silently before expiration
 * - Session timeout forces logout automatically
 * - Role checks happen on both client and server
 */

export default {};
