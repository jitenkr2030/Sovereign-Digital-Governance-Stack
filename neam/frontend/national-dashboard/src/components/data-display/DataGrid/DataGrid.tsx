import React, { useState, useCallback } from 'react';
import type { Column, SortDirection, PaginationState } from './types';
import { LoadingSpinner } from '../ui/LoadingSpinner';
import { EmptyState } from '../ui/EmptyState';
import './DataGrid.css';

interface DataGridProps<T> {
  /** Unique identifier for the data key field */
  dataKey: keyof T;
  /** Array of column definitions */
  columns: Column<T>[];
  /** Data to display */
  data: T[];
  /** Total number of records (for server-side pagination) */
  totalCount?: number;
  /** Current pagination state */
  pagination?: PaginationState;
  /** Callback when pagination changes */
  onPaginationChange?: (pagination: PaginationState) => void;
  /** Current sort state */
  sort?: { column: keyof T; direction: SortDirection };
  /** Callback when sort changes */
  onSortChange?: (column: keyof T, direction: SortDirection) => void;
  /** Is data currently loading */
  isLoading?: boolean;
  /** Error message to display */
  error?: string | null;
  /** Callback when a row is clicked */
  onRowClick?: (row: T) => void;
  /** Custom empty state message */
  emptyMessage?: string;
  /** Row clickable indicator */
  clickableRows?: boolean;
  /** Compact mode for dense displays */
  compact?: boolean;
  /** Sticky header */
  stickyHeader?: boolean;
  /** Checkbox column configuration */
  selectable?: boolean;
  /** Selected row IDs */
  selectedRows?: (string | number)[];
  /** Callback when selection changes */
  onSelectionChange?: (selected: (string | number)[]) => void;
  /** Maximum height of the grid */
  maxHeight?: string;
  /** Custom class name */
  className?: string;
}

/**
 * DataGrid Component
 * 
 * A powerful data table component with server-side sorting, filtering, and pagination.
 * Supports row selection, click actions, and customizable cell rendering.
 */
export function DataGrid<T>({
  dataKey,
  columns,
  data,
  totalCount = data.length,
  pagination = { page: 1, pageSize: 10 },
  onPaginationChange,
  sort,
  onSortChange,
  isLoading = false,
  error = null,
  onRowClick,
  emptyMessage = 'No data available',
  clickableRows = false,
  compact = false,
  stickyHeader = true,
  selectable = false,
  selectedRows = [],
  onSelectionChange,
  maxHeight = '600px',
  className = '',
}: DataGridProps<T>) {
  const [internalSort, setInternalSort] = useState<{ column: keyof T; direction: SortDirection } | undefined>(sort);

  const handleSort = useCallback((column: keyof T) => {
    if (!onSortChange && !onPaginationChange) return;

    const currentSort = sort || internalSort;
    const newDirection: SortDirection = 
      currentSort?.column === column && currentSort.direction === 'asc' ? 'desc' : 'asc';
    
    if (onSortChange) {
      onSortChange(column, newDirection);
    } else {
      setInternalSort({ column, direction: newDirection });
    }
  }, [sort, internalSort, onSortChange, onPaginationChange]);

  const handleSelectAll = useCallback(() => {
    if (!selectable || !onSelectionChange) return;
    
    if (selectedRows.length === data.length) {
      onSelectionChange([]);
    } else {
      onSelectionChange(data.map(row => row[dataKey] as string | number));
    }
  }, [selectable, onSelectionChange, selectedRows, data, dataKey]);

  const handleSelectRow = useCallback((id: string | number) => {
    if (!selectable || !onSelectionChange) return;
    
    if (selectedRows.includes(id)) {
      onSelectionChange(selectedRows.filter(rowId => rowId !== id));
    } else {
      onSelectionChange([...selectedRows, id]);
    }
  }, [selectable, onSelectionChange, selectedRows]);

  const handlePageChange = useCallback((newPage: number) => {
    if (onPaginationChange) {
      onPaginationChange({ ...pagination, page: newPage });
    }
  }, [pagination, onPaginationChange]);

  const handlePageSizeChange = useCallback((newPageSize: number) => {
    if (onPaginationChange) {
      onPaginationChange({ page: 1, pageSize: newPageSize });
    }
  }, [onPaginationChange]);

  const allSelected = selectable && data.length > 0 && selectedRows.length === data.length;
  const someSelected = selectable && selectedRows.length > 0 && selectedRows.length < data.length;

  return (
    <div className={`data-grid-container ${className}`} style={{ maxHeight }}>
      {/* Loading Overlay */}
      {isLoading && (
        <div className="data-grid-loading-overlay">
          <LoadingSpinner size="lg" label="Loading data..." />
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="data-grid-error">
          <EmptyState
            icon="error"
            title="Error loading data"
            description={error}
          />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && !error && data.length === 0 && (
        <div className="data-grid-empty">
          <EmptyState
            icon="empty"
            title="No data found"
            description={emptyMessage}
          />
        </div>
      )}

      {/* Data Table */}
      {!error && data.length > 0 && (
        <div className="data-grid-wrapper">
          <table className={`data-grid ${compact ? 'data-grid-compact' : ''} ${clickableRows ? 'data-grid-clickable' : ''}`}>
            <thead className={stickyHeader ? 'data-grid-header-sticky' : ''}>
              <tr>
                {/* Checkbox Column */}
                {selectable && (
                  <th className="data-grid-cell data-grid-cell-checkbox">
                    <input
                      type="checkbox"
                      checked={allSelected}
                      indeterminate={someSelected}
                      onChange={handleSelectAll}
                      aria-label="Select all rows"
                    />
                  </th>
                )}
                
                {/* Column Headers */}
                {columns.map((column) => {
                  const sortKey = column.sortKey || column.key;
                  const isSorted = (sort || internalSort)?.column === sortKey;
                  const sortDirection = (sort || internalSort)?.direction;
                  const sortable = column.sortable !== false;

                  return (
                    <th
                      key={String(column.key)}
                      className={`data-grid-cell data-grid-cell-header ${
                        sortable ? 'data-grid-cell-sortable' : ''
                      } ${isSorted ? 'data-grid-cell-sorted' : ''}`}
                      style={{ width: column.width, textAlign: column.align }}
                      onClick={() => sortable && handleSort(sortKey)}
                    >
                      <div className="data-grid-header-content">
                        {column.header}
                        {isSorted && (
                          <span className="data-grid-sort-icon">
                            {sortDirection === 'asc' ? '↑' : '↓'}
                          </span>
                        )}
                      </div>
                    </th>
                  );
                })}
              </tr>
            </thead>
            <tbody className="data-grid-body">
              {data.map((row, rowIndex) => {
                const rowId = row[dataKey] as string | number;
                const isSelected = selectedRows.includes(rowId);
                const rowClickable = clickableRows || onRowClick;

                return (
                  <tr
                    key={String(rowId)}
                    className={`data-grid-row ${isSelected ? 'data-grid-row-selected' : ''} ${
                      rowClickable ? 'data-grid-row-clickable' : ''
                    }`}
                    onClick={() => rowClickable && onRowClick?.(row)}
                  >
                    {/* Checkbox */}
                    {selectable && (
                      <td className="data-grid-cell data-grid-cell-checkbox">
                        <input
                          type="checkbox"
                          checked={isSelected}
                          onChange={() => handleSelectRow(rowId)}
                          onClick={(e) => e.stopPropagation()}
                          aria-label={`Select row ${rowId}`}
                        />
                      </td>
                    )}

                    {/* Data Cells */}
                    {columns.map((column) => {
                      const value = row[column.key];
                      const cellContent = column.render
                        ? column.render(value, row, rowIndex)
                        : String(value ?? '');

                      return (
                        <td
                          key={String(column.key)}
                          className="data-grid-cell"
                          style={{ textAlign: column.align }}
                        >
                          {cellContent}
                        </td>
                      );
                    })}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Pagination */}
      {(onPaginationChange || totalCount > pagination.pageSize) && (
        <div className="data-grid-pagination">
          <div className="data-grid-pagination-info">
            Showing {(pagination.page - 1) * pagination.pageSize + 1} to{' '}
            {Math.min(pagination.page * pagination.pageSize, totalCount)} of{' '}
            {totalCount} results
          </div>
          
          <div className="data-grid-pagination-controls">
            {/* Page Size Selector */}
            <select
              className="data-grid-page-size-select"
              value={pagination.pageSize}
              onChange={(e) => handlePageSizeChange(Number(e.target.value))}
            >
              {[10, 25, 50, 100].map((size) => (
                <option key={size} value={size}>
                  {size} per page
                </option>
              ))}
            </select>

            {/* Page Navigation */}
            <div className="data-grid-pagination-buttons">
              <button
                className="data-grid-pagination-btn"
                onClick={() => handlePageChange(1)}
                disabled={pagination.page === 1}
                aria-label="First page"
              >
                ««
              </button>
              <button
                className="data-grid-pagination-btn"
                onClick={() => handlePageChange(pagination.page - 1)}
                disabled={pagination.page === 1}
                aria-label="Previous page"
              >
                «
              </button>
              
              {/* Page Numbers */}
              {generatePageNumbers(pagination.page, totalCount, pagination.pageSize).map((page, idx) => (
                <button
                  key={idx}
                  className={`data-grid-pagination-btn ${
                    page === pagination.page ? 'data-grid-pagination-btn-active' : ''
                  } ${page === '...' ? 'data-grid-pagination-btn-ellipsis' : ''}`}
                  onClick={() => typeof page === 'number' && handlePageChange(page)}
                  disabled={page === '...'}
                  aria-label={typeof page === 'number' ? `Page ${page}` : 'Ellipsis'}
                >
                  {page}
                </button>
              ))}
              
              <button
                className="data-grid-pagination-btn"
                onClick={() => handlePageChange(pagination.page + 1)}
                disabled={pagination.page * pagination.pageSize >= totalCount}
                aria-label="Next page"
              >
                »
              </button>
              <button
                className="data-grid-pagination-btn"
                onClick={() => handlePageChange(Math.ceil(totalCount / pagination.pageSize))}
                disabled={pagination.page * pagination.pageSize >= totalCount}
                aria-label="Last page"
              >
              >>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

// Helper function to generate pagination page numbers
function generatePageNumbers(
  currentPage: number,
  totalCount: number,
  pageSize: number
): (number | '...')[] {
  const totalPages = Math.ceil(totalCount / pageSize);
  const delta = 2;
  const range: (number | '...')[] = [];
  const rangeWithDots: (number | '...')[] = [];

  if (totalPages <= 7) {
    for (let i = 1; i <= totalPages; i++) {
      range.push(i);
    }
  } else {
    if (currentPage <= delta + 1) {
      for (let i = 1; i <= delta + 2; i++) {
        range.push(i);
      }
      range.push('...');
      range.push(totalPages);
    } else if (currentPage >= totalPages - delta - 1) {
      range.push(1);
      range.push('...');
      for (let i = totalPages - delta - 1; i <= totalPages; i++) {
        range.push(i);
      }
    } else {
      range.push(1);
      range.push('...');
      for (let i = currentPage - delta; i <= currentPage + delta; i++) {
        range.push(i);
      }
      range.push('...');
      range.push(totalPages);
    }
  }

  range.forEach((page) => {
    rangeWithDots.push(page);
  });

  return rangeWithDots;
}

export default DataGrid;
