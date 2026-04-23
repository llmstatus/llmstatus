import { render, screen } from '@testing-library/react'
import { StatusIndicator } from '@/components/ui/StatusIndicator'

describe('StatusIndicator', () => {
  it('should render operational status correctly', () => {
    render(<StatusIndicator status="operational" />)

    const indicator = screen.getByRole('status')
    expect(indicator).toHaveTextContent('Operational')
    expect(indicator).toHaveClass('indicator')
    expect(indicator).toHaveClass('operational')
  })

  it('should render degraded status with warning', () => {
    render(<StatusIndicator status="degraded" />)

    const indicator = screen.getByRole('status')
    expect(indicator).toHaveTextContent('Degraded')
    expect(indicator).toHaveClass('indicator')
    expect(indicator).toHaveClass('degraded')
  })

  it('should render down status with error', () => {
    render(<StatusIndicator status="down" />)

    const indicator = screen.getByRole('status')
    expect(indicator).toHaveTextContent('Down')
    expect(indicator).toHaveClass('indicator')
    expect(indicator).toHaveClass('down')
  })
})