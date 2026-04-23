'use client'

import { useState, useMemo } from 'react'
import { cn } from '@/lib/utils'
import styles from './TimeSeriesChart.module.css'

export interface DataPoint {
  timestamp: string
  value: number
}

interface TimeSeriesChartProps {
  data: DataPoint[]
  title: string
  unit?: string
  loading?: boolean
  className?: string
  height?: number
}

export function TimeSeriesChart({
  data,
  title,
  unit = '',
  loading = false,
  className,
  height = 200,
}: TimeSeriesChartProps) {
  const [tooltip, setTooltip] = useState<{
    visible: boolean
    x: number
    y: number
    content: string
  }>({ visible: false, x: 0, y: 0, content: '' })

  const chartData = useMemo(() => {
    if (data.length === 0) return null

    const values = data.map(d => d.value)
    const minValue = Math.min(...values)
    const maxValue = Math.max(...values)
    const valueRange = maxValue - minValue || 1

    const width = 400
    const chartHeight = height - 40 // Account for padding
    const padding = 20

    const points = data.map((point, index) => {
      const x = padding + (index / (data.length - 1)) * (width - 2 * padding)
      const y = padding + ((maxValue - point.value) / valueRange) * (chartHeight - 2 * padding)
      return { x, y, ...point }
    })

    const pathData = points
      .map((point, index) => `${index === 0 ? 'M' : 'L'} ${point.x} ${point.y}`)
      .join(' ')

    return { points, pathData, width, height: chartHeight, minValue, maxValue }
  }, [data, height])

  const handleMouseMove = (event: React.MouseEvent, point: DataPoint) => {
    const rect = event.currentTarget.getBoundingClientRect()
    setTooltip({
      visible: true,
      x: event.clientX - rect.left,
      y: event.clientY - rect.top,
      content: `${new Date(point.timestamp).toLocaleTimeString()}: ${point.value}${unit}`,
    })
  }

  const handleMouseLeave = () => {
    setTooltip({ visible: false, x: 0, y: 0, content: '' })
  }

  if (loading) {
    return (
      <div className={cn(styles.container, className)}>
        <div className={styles.header}>
          <h3 className={styles.title}>{title}</h3>
        </div>
        <div className={styles.loading} role="status" aria-label="loading">
          <div className={styles.loadingSpinner} />
          Loading chart data...
        </div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className={cn(styles.container, className)}>
        <div className={styles.header}>
          <h3 className={styles.title}>{title}</h3>
        </div>
        <div className={styles.empty}>
          No data available
        </div>
      </div>
    )
  }

  return (
    <div className={cn(styles.container, className)}>
      <div className={styles.header}>
        <h3 className={styles.title}>{title}</h3>
      </div>
      <div className={styles.chartContainer}>
        <svg
          className={styles.chart}
          viewBox={`0 0 ${chartData?.width} ${chartData?.height}`}
          role="img"
          aria-label={`${title} chart`}
          onMouseLeave={handleMouseLeave}
        >
          {chartData && (
            <>
              {/* Chart line */}
              <path
                d={chartData.pathData}
                fill="none"
                stroke="var(--color-status-operational)"
                strokeWidth="2"
              />

              {/* Data points */}
              {chartData.points.map((point, index) => (
                <circle
                  key={index}
                  cx={point.x}
                  cy={point.y}
                  r="4"
                  fill="var(--color-status-operational)"
                  stroke="var(--color-bg-primary)"
                  strokeWidth="2"
                  style={{ cursor: 'pointer' }}
                  onMouseMove={(e) => handleMouseMove(e, point)}
                />
              ))}
            </>
          )}
        </svg>

        {/* Tooltip */}
        {tooltip.visible && (
          <div
            className={styles.tooltip}
            style={{
              left: tooltip.x + 10,
              top: tooltip.y - 10,
            }}
          >
            {tooltip.content}
          </div>
        )}
      </div>
    </div>
  )
}