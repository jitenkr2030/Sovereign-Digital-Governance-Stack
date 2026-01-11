import React from 'react';

interface PaginationProps {
  currentPage: number;
  totalPages: number;
  totalItems: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (size: number) => void;
  showPageSizeSelector?: boolean;
  className?: string;
}

export const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  totalPages,
  totalItems,
  pageSize,
  onPageChange,
  onPageSizeChange,
  showPageSizeSelector = true,
  className = '',
}) => {
  const pageSizeOptions = [10, 25, 50, 100];

  const startItem = (currentPage - 1) * pageSize + 1;
  const endItem = Math.min(currentPage * pageSize, totalItems);

  const pages = React.useMemo(() => {
    const result: (number | string)[] = [];
    const showPages = 5;

    if (totalPages <= showPages + 2) {
      for (let i = 1; i <= totalPages; i++) {
        result.push(i);
      }
    } else {
      result.push(1);

      const startPage = Math.max(2, currentPage - Math.floor(showPages / 2));
      const endPage = Math.min(totalPages - 1, currentPage + Math.floor(showPages / 2));

      if (startPage > 2) {
        result.push('...');
      }

      for (let i = startPage; i <= endPage; i++) {
        result.push(i);
      }

      if (endPage < totalPages - 1) {
        result.push('...');
      }

      result.push(totalPages);
    }

    return result;
  }, [currentPage, totalPages]);

  const handlePageClick = (page: number | string) => {
    if (typeof page === 'number' && page !== currentPage && page >= 1 && page <= totalPages) {
      onPageChange(page);
    }
  };

  return (
    <div className={`flex flex-col sm:flex-row items-center justify-between gap-4 ${className}`}>
      <div className="text-sm text-slate-600">
        Showing <span className="font-medium">{startItem}</span> to{' '}
        <span className="font-medium">{endItem}</span> of{' '}
        <span className="font-medium">{totalItems}</span> results
      </div>

      <div className="flex items-center gap-4">
        {showPageSizeSelector && onPageSizeChange && (
          <div className="flex items-center gap-2">
            <span className="text-sm text-slate-600">Show</span>
            <select
              value={pageSize}
              onChange={(e) => onPageSizeChange(Number(e.target.value))}
              className="px-2 py-1 border border-slate-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {pageSizeOptions.map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
            <span className="text-sm text-slate-600">per page</span>
          </div>
        )}

        <nav className="flex items-center gap-1">
          <button
            onClick={() => onPageChange(currentPage - 1)}
            disabled={currentPage === 1}
            className="p-2 border border-slate-300 rounded-lg text-slate-600 hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
            aria-label="Previous page"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
            </svg>
          </button>

          {pages.map((page, index) =>
            typeof page === 'string' ? (
              <span
                key={`ellipsis-${index}`}
                className="px-3 py-2 text-slate-500"
              >
                ...
              </span>
            ) : (
              <button
                key={page}
                onClick={() => handlePageClick(page)}
                className={`
                  min-w-[40px] h-10 px-3 rounded-lg text-sm font-medium
                  ${page === currentPage
                    ? 'bg-blue-500 text-white'
                    : 'border border-slate-300 text-slate-600 hover:bg-slate-100'
                  }
                `}
                aria-label={`Page ${page}`}
                aria-current={page === currentPage ? 'page' : undefined}
              >
                {page}
              </button>
            )
          )}

          <button
            onClick={() => onPageChange(currentPage + 1)}
            disabled={currentPage === totalPages}
            className="p-2 border border-slate-300 rounded-lg text-slate-600 hover:bg-slate-100 disabled:opacity-50 disabled:cursor-not-allowed"
            aria-label="Next page"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </button>
        </nav>
      </div>
    </div>
  );
};

interface PageInfo {
  page: number;
  pageSize: number;
  totalItems: number;
  totalPages: number;
}

export const usePagination = (initialPageInfo: PageInfo) => {
  const [pageInfo, setPageInfo] = React.useState<PageInfo>(initialPageInfo);

  const setPage = (page: number) => {
    setPageInfo((prev) => ({ ...prev, page }));
  };

  const setPageSize = (pageSize: number) => {
    setPageInfo((prev) => ({
      ...prev,
      pageSize,
      page: 1,
    }));
  };

  const updateFromResponse = (response: { page: number; pageSize: number; totalItems: number; totalPages: number }) => {
    setPageInfo((prev) => ({
      ...prev,
      page: response.page,
      pageSize: response.pageSize,
      totalItems: response.totalItems,
      totalPages: response.totalPages,
    }));
  };

  return {
    currentPage: pageInfo.page,
    pageSize: pageInfo.pageSize,
    totalItems: pageInfo.totalItems,
    totalPages: pageInfo.totalPages,
    setPage,
    setPageSize,
    updateFromResponse,
  };
};

export default Pagination;
