import { cn } from '@/lib/utils'
import styles from './StatusIndicator.module.css'

export type Status = 'operational' | 'degraded' | 'down'

interface StatusIndicatorProps {
  status: Status
  className?: string
  showText?: boolean
  size?: 'sm' | 'md' | 'lg'
}

const statusLabels: Record<Status, string> = {
  operational: 'Operational',
  degraded: 'Degraded',
  down: 'Down',
}

export function StatusIndicator({
  status,
  className,
  showText = true,
  size = 'md'
}: StatusIndicatorProps) {
  return (
    <div
      role="status"
      aria-label={`Status: ${statusLabels[status]}`}
      className={cn(
        styles.indicator,
        styles[status],
        className
      )}
    >
      <div className={styles.dot} />
      {showText && (
        <span>{statusLabels[status]}</span>
      )}
    </div>
  )
}