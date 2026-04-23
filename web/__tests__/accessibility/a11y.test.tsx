import { render } from '@testing-library/react'
import { axe, toHaveNoViolations } from 'jest-axe'
import { StatusIndicator } from '@/components/ui/StatusIndicator'
import { MetricCard } from '@/components/ui/MetricCard'

expect.extend(toHaveNoViolations)

describe('Accessibility', () => {
  it('StatusIndicator should have no accessibility violations', async () => {
    const { container } = render(<StatusIndicator status="operational" />)
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })

  it('MetricCard should have no accessibility violations', async () => {
    const { container } = render(
      <MetricCard title="Uptime" value="99.9%" description="Last 30 days" />
    )
    const results = await axe(container)
    expect(results).toHaveNoViolations()
  })
})