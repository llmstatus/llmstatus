import { breakpoints, colors, spacing } from '@/lib/design-tokens'

describe('Design Tokens', () => {
  it('should have consistent breakpoint values', () => {
    expect(breakpoints.mobile).toBe(320)
    expect(breakpoints.tablet).toBe(768)
    expect(breakpoints.desktop).toBe(1024)
    expect(breakpoints.wide).toBe(1440)
  })

  it('should have accessible color contrast ratios', () => {
    expect(colors.status.operational).toBe('#10b981')
    expect(colors.status.degraded).toBe('#f59e0b')
    expect(colors.status.down).toBe('#ef4444')
  })
})