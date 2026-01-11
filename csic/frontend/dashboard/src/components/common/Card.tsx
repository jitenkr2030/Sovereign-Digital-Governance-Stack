import { ReactNode } from 'react';
import clsx from 'clsx';

interface CardProps {
  children: ReactNode;
  className?: string;
  hover?: boolean;
  padding?: 'none' | 'sm' | 'md' | 'lg';
}

interface CardHeaderProps {
  children: ReactNode;
  className?: string;
  action?: ReactNode;
}

interface CardBodyProps {
  children: ReactNode;
  className?: string;
}

interface CardFooterProps {
  children: ReactNode;
  className?: string;
}

export function Card({
  children,
  className,
  hover = false,
  padding = 'md',
}: CardProps) {
  const paddingStyles = {
    none: '',
    sm: 'p-4',
    md: 'p-6',
    lg: 'p-8',
  };

  return (
    <div
      className={clsx(
        'bg-white dark:bg-dark-900 rounded-xl border border-dark-200 dark:border-dark-700 shadow-sm',
        hover && 'hover:shadow-md transition-shadow duration-200',
        paddingStyles[padding],
        className
      )}
    >
      {children}
    </div>
  );
}

export function CardHeader({ children, className, action }: CardHeaderProps) {
  return (
    <div
      className={clsx(
        'flex items-center justify-between mb-4 pb-4 border-b border-dark-200 dark:border-dark-700',
        className
      )}
    >
      <div>{children}</div>
      {action && <div>{action}</div>}
    </div>
  );
}

export function CardTitle({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <h3 className={clsx('text-lg font-semibold text-dark-900 dark:text-dark-50', className)}>
      {children}
    </h3>
  );
}

export function CardDescription({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <p className={clsx('text-sm text-dark-500 dark:text-dark-400 mt-1', className)}>
      {children}
    </p>
  );
}

export function CardBody({ children, className }: CardBodyProps) {
  return <div className={className}>{children}</div>;
}

export function CardFooter({ children, className }: CardFooterProps) {
  return (
    <div
      className={clsx(
        'mt-4 pt-4 border-t border-dark-200 dark:border-dark-700 flex items-center justify-end gap-2',
        className
      )}
    >
      {children}
    </div>
  );
}

// Stats Card Component
interface StatsCardProps {
  title: string;
  value: string | number;
  change?: {
    value: number;
    type: 'increase' | 'decrease' | 'neutral';
  };
  icon?: ReactNode;
  iconBg?: string;
}

export function StatsCard({
  title,
  value,
  change,
  icon,
  iconBg = 'bg-primary-100 dark:bg-primary-900/30',
}: StatsCardProps) {
  return (
    <Card>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-sm font-medium text-dark-500 dark:text-dark-400">{title}</p>
          <p className="text-3xl font-bold text-dark-900 dark:text-dark-50 mt-2">{value}</p>
          {change && (
            <p
              className={clsx(
                'text-sm mt-2 flex items-center gap-1',
                change.type === 'increase' && 'text-green-600 dark:text-green-400',
                change.type === 'decrease' && 'text-red-600 dark:text-red-400',
                change.type === 'neutral' && 'text-dark-500 dark:text-dark-400'
              )}
            >
              {change.type === 'increase' && '↑'}
              {change.type === 'decrease' && '↓'}
              {change.value}% from last period
            </p>
          )}
        </div>
        {icon && (
          <div className={clsx('p-3 rounded-lg', iconBg)}>
            {icon}
          </div>
        )}
      </div>
    </Card>
  );
}
