import { useState, useEffect, useCallback, useRef } from 'react';
import { FilterOptions, SortOptions, PaginatedResponse } from '../types';

interface UseDataFetchOptions<T> {
  immediate?: boolean;
  deps?: unknown[];
}

interface UseDataFetchResult<T> {
  data: T[];
  isLoading: boolean;
  error: Error | null;
  refetch: () => Promise<void>;
  loadMore: () => Promise<void>;
  hasMore: boolean;
}

export function useDataFetch<T>(
  fetchFn: (params: { page: number; filters?: FilterOptions; sort?: SortOptions }) => Promise<PaginatedResponse<T>>,
  options: UseDataFetchOptions<T> = {}
): UseDataFetchResult<T> & { filters: FilterOptions; setFilters: (filters: FilterOptions) => void; sort: SortOptions; setSort: (sort: SortOptions) => void } {
  const { immediate = true, deps = [] } = options;
  const [data, setData] = useState<T[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);
  const [filters, setFilters] = useState<FilterOptions>({});
  const [sort, setSort] = useState<SortOptions>({ field: 'createdAt', direction: 'desc' });
  const mountedRef = useRef(true);

  const fetchData = useCallback(async (resetData = false) => {
    if (!mountedRef.current) return;
    
    setIsLoading(true);
    setError(null);

    try {
      const currentPage = resetData ? 1 : page;
      const response = await fetchFn({ page: currentPage, filters, sort });

      if (!mountedRef.current) return;

      if (resetData) {
        setData(response.items);
      } else {
        setData((prev) => [...prev, ...response.items]);
      }

      setHasMore(response.meta.page < response.meta.totalPages);
      setPage(currentPage + 1);
    } catch (err) {
      if (!mountedRef.current) return;
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      if (mountedRef.current) {
        setIsLoading(false);
      }
    }
  }, [fetchFn, filters, sort, page]);

  useEffect(() => {
    mountedRef.current = true;
    if (immediate) {
      fetchData(true);
    }
    return () => {
      mountedRef.current = false;
    };
  }, [immediate, ...deps]);

  const loadMore = useCallback(async () => {
    if (!isLoading && hasMore) {
      await fetchData(false);
    }
  }, [isLoading, hasMore, fetchData]);

  const updateFilters = useCallback((newFilters: FilterOptions) => {
    setFilters(newFilters);
    setPage(1);
    setHasMore(true);
  }, []);

  const updateSort = useCallback((newSort: SortOptions) => {
    setSort(newSort);
    setPage(1);
    setHasMore(true);
  }, []);

  return {
    data,
    isLoading,
    error,
    refetch: () => fetchData(true),
    loadMore,
    hasMore,
    filters,
    setFilters: updateFilters,
    sort,
    setSort: updateSort,
  };
}

export function useAsync<T>(
  asyncFn: () => Promise<T>,
  immediate = true
): { data: T | null; isLoading: boolean; error: Error | null; execute: () => Promise<void> } {
  const [data, setData] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const mountedRef = useRef(true);

  const execute = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const result = await asyncFn();
      if (!mountedRef.current) return;
      setData(result);
    } catch (err) {
      if (!mountedRef.current) return;
      setError(err instanceof Error ? err : new Error('An error occurred'));
    } finally {
      if (mountedRef.current) {
        setIsLoading(false);
      }
    }
  }, [asyncFn]);

  useEffect(() => {
    mountedRef.current = true;
    if (immediate) {
      execute();
    }
    return () => {
      mountedRef.current = false;
    };
  }, [execute, immediate]);

  return { data, isLoading, error, execute };
}

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState<T>(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(timer);
    };
  }, [value, delay]);

  return debouncedValue;
}

export function useLocalStorage<T>(key: string, initialValue: T): [T, (value: T | ((val: T) => T)) => void] {
  const [storedValue, setStoredValue] = useState<T>(() => {
    try {
      const item = window.localStorage.getItem(key);
      return item ? JSON.parse(item) : initialValue;
    } catch {
      return initialValue;
    }
  });

  const setValue = (value: T | ((val: T) => T)) => {
    try {
      const valueToStore = value instanceof Function ? value(storedValue) : value;
      setStoredValue(valueToStore);
      window.localStorage.setItem(key, JSON.stringify(valueToStore));
    } catch (error) {
      console.error('Error saving to localStorage:', error);
    }
  };

  return [storedValue, setValue];
}
