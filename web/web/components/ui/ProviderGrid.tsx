import { cn } from '@/lib/utils'
import { StatusIndicator, type Status } from './StatusIndicator'
import styles from './ProviderGrid.module.css'

export interface Provider {
  id: string
  name: string
  status: Status
  lastUpdated?: string
}

interface ProviderGridProps {
  providers: Provider[]
  loading?: boolean
  onProviderClick?: (provider: Provider) => void
  className?: string
}

function ProviderCard({ provider, onClick }: { provider: Provider; onClick?: (provider: Provider) => void }) {
  const handleClick = () => onClick?.(provider)
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault()
      handleClick()
    }
  }

  return (
    <div
      className={styles.providerCard}
      onClick={handleClick}
      onKeyDown={handleKeyDown}
      role="button"
      tabIndex={0}
      aria-label={`${provider.name} - ${provider.status}`}
    >
      <div className={styles.providerHeader}>
        <h3 className={styles.providerName}>{provider.name}</h3>
        <StatusIndicator status={provider.status} showText={false} />
      </div>

      <div className={styles.providerMeta}>
        {provider.lastUpdated && (
          <p className={styles.lastUpdated}>
            Updated {new Date(provider.lastUpdated).toLocaleString()}
          </p>
        )}
      </div>
    </div>
  )
}

function SkeletonCard() {
  return <div className={styles.skeleton} role="status" aria-label="loading" />
}

function EmptyState() {
  return (
    <div className={styles.emptyState}>
      <svg className={styles.emptyIcon} fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
      </svg>
      <h3 className={styles.emptyTitle}>No providers found</h3>
      <p className={styles.emptyDescription}>
        Provider data will appear here when available
      </p>
    </div>
  )
}

export function ProviderGrid({
  providers,
  loading = false,
  onProviderClick,
  className
}: ProviderGridProps) {
  if (loading) {
    return (
      <div className={cn(styles.grid, className)}>
        {Array.from({ length: 8 }, (_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  if (providers.length === 0) {
    return (
      <div className={cn(styles.grid, className)}>
        <EmptyState />
      </div>
    )
  }

  return (
    <div className={cn(styles.grid, className)}>
      {providers.map((provider) => (
        <ProviderCard
          key={provider.id}
          provider={provider}
          onClick={onProviderClick}
        />
      ))}
    </div>
  )
}