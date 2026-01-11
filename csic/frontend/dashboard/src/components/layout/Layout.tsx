import { ReactNode } from 'react';
import Sidebar from './Sidebar';
import clsx from 'clsx';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  return (
    <div className="min-h-screen bg-dark-50 dark:bg-dark-950">
      <Sidebar />
      <main
        className={clsx(
          'min-h-screen transition-all duration-300',
          'lg:ml-64 ml-16'
        )}
      >
        <div className="p-4 lg:p-6">
          {children}
        </div>
      </main>
    </div>
  );
}
