import { useState } from 'react';
import { useLocation } from 'react-router-dom';
import { Search, Bell, Menu, X } from 'lucide-react';
import clsx from 'clsx';

interface HeaderProps {
  title?: string;
  description?: string;
  showSearch?: boolean;
  onSearch?: (query: string) => void;
  actions?: React.ReactNode;
}

export default function Header({
  title,
  description,
  showSearch = false,
  onSearch,
  actions,
}: HeaderProps) {
  const location = useLocation();
  const [searchQuery, setSearchQuery] = useState('');
  const [showMobileMenu, setShowMobileMenu] = useState(false);

  const getPageTitle = () => {
    if (title) return { title, description };

    const titles: Record<string, { title: string; description?: string }> = {
      '/': {
        title: 'Dashboard',
        description: 'Overview of regulatory compliance and monitoring',
      },
      '/licenses': {
        title: 'License Management',
        description: 'Manage VASP/CASP licenses and applications',
      },
      '/energy': {
        title: 'Energy Analytics',
        description: 'Monitor energy consumption and grid performance',
      },
      '/reports': {
        title: 'Regulatory Reports',
        description: 'Generate and manage compliance reports',
      },
      '/settings': {
        title: 'Settings',
        description: 'Configure system preferences and integrations',
      },
    };

    return titles[location.pathname] || { title: 'CSIC Dashboard', description: '' };
  };

  const page = getPageTitle();

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    onSearch?.(searchQuery);
  };

  return (
    <header className="sticky top-0 z-30 bg-white/80 dark:bg-dark-900/80 backdrop-blur-lg border-b border-dark-200 dark:border-dark-700">
      <div className="flex items-center justify-between h-16 px-4 lg:px-6">
        {/* Mobile menu button */}
        <button
          onClick={() => setShowMobileMenu(!showMobileMenu)}
          className="lg:hidden p-2 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-800"
        >
          {showMobileMenu ? <X size={24} /> : <Menu size={24} />}
        </button>

        {/* Page title - Desktop */}
        <div className="hidden lg:block">
          <h1 className="text-xl font-semibold text-dark-900 dark:text-dark-50">
            {page.title}
          </h1>
          {page.description && (
            <p className="text-sm text-dark-500 dark:text-dark-400">
              {page.description}
            </p>
          )}
        </div>

        {/* Search */}
        {showSearch && (
          <form onSubmit={handleSearch} className="flex-1 max-w-md mx-4 lg:mx-8">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-dark-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search..."
                className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800 text-dark-900 dark:text-dark-100 placeholder-dark-400 focus:outline-none focus:ring-2 focus:ring-primary-500"
              />
            </div>
          </form>
        )}

        {/* Mobile page title */}
        <div className="lg:hidden">
          <h1 className="text-lg font-semibold text-dark-900 dark:text-dark-50">
            {page.title}
          </h1>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {actions}

          {/* Notifications */}
          <button className="relative p-2 rounded-lg hover:bg-dark-100 dark:hover:bg-dark-800 transition-colors">
            <Bell className="w-5 h-5 text-dark-600 dark:text-dark-400" />
            <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
          </button>
        </div>
      </div>

      {/* Mobile menu */}
      {showMobileMenu && (
        <div className="lg:hidden border-t border-dark-200 dark:border-dark-700 py-4 px-4">
          <form onSubmit={handleSearch} className="mb-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-dark-400" />
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search..."
                className="w-full pl-10 pr-4 py-2 rounded-lg border border-dark-300 dark:border-dark-600 bg-white dark:bg-dark-800"
              />
            </div>
          </form>
        </div>
      )}
    </header>
  );
}
