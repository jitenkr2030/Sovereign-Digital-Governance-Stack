import type { ReactNode } from 'react';

export type ColumnAlign = 'left' | 'center' | 'right';
export type SortDirection = 'asc' | 'desc';

export interface Column<T> {
  /** Unique key for the column (matches data field) */
  key: keyof T;
  /** Column header display text */
  header: string;
  /** Custom render function for cell content */
  render?: (value: unknown, row: T, index: number) => ReactNode;
  /** Width of the column (CSS value) */
  width?: string;
  /** Text alignment */
  align?: ColumnAlign;
  /** Whether this column is sortable (default: true) */
  sortable?: boolean;
  /** Custom sort key (if different from key) */
  sortKey?: keyof T;
  /** Custom class name for the cell */
  cellClassName?: string;
}

export interface PaginationState {
  /** Current page number (1-indexed) */
  page: number;
  /** Number of records per page */
  pageSize: number;
}

export interface FilterState {
  /** Search query string */
  search?: string;
  /** Column-specific filters */
  filters?: Record<string, string>;
}

export interface DataGridState<T> {
  /** Current pagination */
  pagination: PaginationState;
  /** Current sort */
  sort?: { column: keyof T; direction: SortDirection };
  /** Current filters */
  filters: FilterState;
}

export interface DataGridServerConfig {
  /** API endpoint for fetching data */
  endpoint: string;
  /** Additional query parameters */
  params?: Record<string, string>;
  /** Request method (default: GET) */
  method?: 'GET' | 'POST';
  /** Custom fetch function */
  fetch?: typeof fetch;
}

export interface DataGridFeatures {
  /** Enable server-side pagination */
  serverPagination?: boolean;
  /** Enable server-side sorting */
  serverSorting?: boolean;
  /** Enable server-side filtering */
  serverFiltering?: boolean;
  /** Enable row selection */
  selection?: boolean;
  /** Enable row expansion (for details) */
  expandable?: boolean;
  /** Enable column resizing */
  resizable?: boolean;
  /** Enable column reordering */
  reorderable?: boolean;
  /** Enable frozen columns */
  frozenColumns?: (keyof T)[];
}

export interface DataGridEventHandlers<T> {
  /** Called when pagination changes */
  onPaginationChange?: (pagination: PaginationState) => void;
  /** Called when sort changes */
  onSortChange?: (column: keyof T, direction: SortDirection) => void;
  /** Called when filters change */
  onFilterChange?: (filters: FilterState) => void;
  /** Called when a row is clicked */
  onRowClick?: (row: T) => void;
  /** Called when selection changes */
  onSelectionChange?: (selected: (string | number)[]) => void;
  /** Called when a row is expanded */
  onRowExpand?: (row: T, expanded: boolean) => void;
  /** Called on data fetch error */
  onError?: (error: Error) => void;
}

export type DataGridProps<T> = {
  /** Unique identifier field for rows */
  dataKey: keyof T;
  /** Column definitions */
  columns: Column<T>[];
  /** Data to display (for client-side) */
  data?: T[];
  /** Server-side configuration */
  serverConfig?: DataGridServerConfig;
  /** Total record count (for server-side) */
  totalCount?: number;
  /** Current pagination state */
  pagination?: PaginationState;
  /** Current sort state */
  sort?: { column: keyof T; direction: SortDirection };
  /** Current filter state */
  filters?: FilterState;
  /** Loading state */
  isLoading?: boolean;
  /** Error message */
  error?: string | null;
  /** Empty state message */
  emptyMessage?: string;
  /** Custom class name */
  className?: string;
  /** Compact mode */
  compact?: boolean;
  /** Sticky header */
  stickyHeader?: boolean;
  /** Maximum height */
  maxHeight?: string;
  /** Features configuration */
  features?: DataGridFeatures;
} & DataGridEventHandlers<T>;
