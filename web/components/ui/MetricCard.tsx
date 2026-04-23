import { cn } from '@/lib/utils'
import styles from './MetricCard.module.css'

export type TrendDirection = 'up' | 'down' | 'flat'

interface MetricCardProps {
  title: string
  value?: string
  trend?: TrendDirection
  trendValue?: string
  description?: string
  loading?: boolean
  className?: string
  onClick?: () => void
}

const TrendIcon = ({ direction }: { direction: TrendDirection }) => {
  switch (direction) {
    case 'up':
      return (
        <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor" role="img" aria-label="trend up">
          <path d="M3.5 8.5L6 6l2.5 2.5L10 7V9.5H7.5L9 8l-2.5-2.5L4 8 3.5 8.5z"/>
        </svg>
      )
    case 'down':
      return (
        <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor" role="img" aria-label="trend down">
          <path d="M8.5 3.5L6 6 3.5 3.5 2 4.5V2H4.5L3 3.5 5.5 6 8 3.5l.5.5z"/>
        </svg>
      )
    case 'flat':
      return (
        <svg width="12" height="12" viewBox="0 0 12 12" fill="currentColor" role="img" aria-label="trend flat">
          <path d="M2 6h8M6 6h0"/>
        </svg>
      )
  }
}

export function MetricCard({
  title,
  value,
  trend,
  trendValue,
  description,
  loading = false,
  className,
  onClick,
}: MetricCardProps) {
  if (loading) {
    return (
      <div className={cn(styles.card, styles.loading, className)} role="status" aria-label="loading">
        <div className={styles.header}>
          <h3 className={styles.title}>{title}</h3>
        </div>
        <div className={cn(styles.skeleton, styles.skeletonValue)} />
        <div className={cn(styles.skeleton, styles.skeletonDescription)} />
      </div>
    )
  }

  return (
    <div
      className={cn(styles.card, className)}
      onClick={onClick}
      role={onClick ? 'button' : undefined}
      tabIndex={onClick ? 0 : undefined}
    >
      <div className={styles.header}>
        <h3 className={styles.title}>{title}</h3>
        {trend && (
          <div className={cn(
            styles.trend,
            styles[`trend${trend.charAt(0).toUpperCase() + trend.slice(1)}`]
          )}>
            <TrendIcon direction={trend} />
            {trendValue && <span>{trendValue}</span>}
          </div>
        )}
      </div>

      {value && <div className={styles.value}>{value}</div>}
      {description && <p className={styles.description}>{description}</p>}
    </div>
  )
}