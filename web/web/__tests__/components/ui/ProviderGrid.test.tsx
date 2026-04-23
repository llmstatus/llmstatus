import { render, screen } from '@testing-library/react'
import { ProviderGrid } from '@/components/ui/ProviderGrid'

const mockProviders = [
  { id: 'openai', name: 'OpenAI', status: 'operational' as const },
  { id: 'anthropic', name: 'Anthropic', status: 'degraded' as const },
  { id: 'google', name: 'Google', status: 'down' as const },
]

describe('ProviderGrid', () => {
  it('should render providers in grid layout', () => {
    render(<ProviderGrid providers={mockProviders} />)

    expect(screen.getByText('OpenAI')).toBeInTheDocument()
    expect(screen.getByText('Anthropic')).toBeInTheDocument()
    expect(screen.getByText('Google')).toBeInTheDocument()
  })

  it('should render loading state', () => {
    render(<ProviderGrid providers={[]} loading />)

    expect(screen.getAllByRole('status')).toHaveLength(8) // Default skeleton count
  })

  it('should handle empty state', () => {
    render(<ProviderGrid providers={[]} />)

    expect(screen.getByText(/no providers/i)).toBeInTheDocument()
  })
})