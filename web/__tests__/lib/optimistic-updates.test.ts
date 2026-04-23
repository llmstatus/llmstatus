import { optimisticUpdate, rollbackUpdate } from '@/lib/optimistic-updates'

describe('optimistic updates', () => {
  it('should apply optimistic update immediately', () => {
    const originalData = [{ id: 'openai', status: 'operational' }]
    const update = { id: 'openai', status: 'degraded' }

    const result = optimisticUpdate(originalData, update)

    expect(result[0].status).toBe('degraded')
    expect(result[0]._optimistic).toBe(true)
  })

  it('should rollback failed optimistic update', () => {
    const optimisticData = [{ id: 'openai', status: 'degraded', _optimistic: true }]
    const originalData = [{ id: 'openai', status: 'operational' }]

    const result = rollbackUpdate(optimisticData, originalData)

    expect(result[0].status).toBe('operational')
    expect(result[0]._optimistic).toBeUndefined()
  })
})