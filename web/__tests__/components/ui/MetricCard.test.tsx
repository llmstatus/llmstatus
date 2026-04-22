import { render, screen } from '@testing-library/react'
import { MetricCard } from '@/components/ui/MetricCard'

describe('MetricCard', () => {
  it('should render uptime metric correctly', () => {
    render(
      <MetricCard
        title="Uptime"
        value="99.9%"
        trend="up"
        description="Last 30 days"
      />
    )

    expect(screen.getByText('Uptime')).toBeInTheDocument()
    expect(screen.getByText('99.9%')).toBeInTheDocument()
    expect(screen.getByText('Last 30 days')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: /trend up/i })).toBeInTheDocument()
  })

  it('should render loading state', () => {
    render(<MetricCard title="Latency" loading />)

    expect(screen.getByText('Latency')).toBeInTheDocument()
    expect(screen.getByRole('status', { name: /loading/i })).toBeInTheDocument()
  })
})