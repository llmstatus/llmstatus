import { render, screen } from '@testing-library/react'
import { TimeSeriesChart } from '@/components/ui/TimeSeriesChart'

const mockData = [
  { timestamp: '2026-04-22T10:00:00Z', value: 150 },
  { timestamp: '2026-04-22T11:00:00Z', value: 200 },
  { timestamp: '2026-04-22T12:00:00Z', value: 180 },
]

describe('TimeSeriesChart', () => {
  it('should render chart with data points', () => {
    render(
      <TimeSeriesChart
        data={mockData}
        title="Response Time"
        unit="ms"
      />
    )

    expect(screen.getByText('Response Time')).toBeInTheDocument()
    expect(screen.getByRole('img', { name: /chart/i })).toBeInTheDocument()
  })

  it('should render loading state', () => {
    render(<TimeSeriesChart data={[]} loading title="Latency" />)

    expect(screen.getByRole('status', { name: /loading/i })).toBeInTheDocument()
  })
})