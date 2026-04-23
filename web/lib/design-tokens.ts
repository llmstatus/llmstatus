export const breakpoints = {
  mobile: 320,
  tablet: 768,
  desktop: 1024,
  wide: 1440,
} as const

export const colors = {
  status: {
    operational: '#10b981', // green-500
    degraded: '#f59e0b',    // amber-500
    down: '#ef4444',        // red-500
  },
  background: {
    primary: '#ffffff',
    secondary: '#f8fafc',   // slate-50
    tertiary: '#f1f5f9',    // slate-100
  },
  text: {
    primary: '#0f172a',     // slate-900
    secondary: '#475569',   // slate-600
    tertiary: '#94a3b8',    // slate-400
  },
  border: {
    light: '#e2e8f0',       // slate-200
    medium: '#cbd5e1',      // slate-300
    dark: '#94a3b8',        // slate-400
  },
} as const

export const spacing = {
  xs: '0.25rem',   // 4px
  sm: '0.5rem',    // 8px
  md: '1rem',      // 16px
  lg: '1.5rem',    // 24px
  xl: '2rem',      // 32px
  '2xl': '3rem',   // 48px
  '3xl': '4rem',   // 64px
} as const

export const typography = {
  fontFamily: {
    sans: ['Inter', 'system-ui', 'sans-serif'],
    mono: ['JetBrains Mono', 'Consolas', 'monospace'],
  },
  fontSize: {
    xs: '0.75rem',   // 12px
    sm: '0.875rem',  // 14px
    base: '1rem',    // 16px
    lg: '1.125rem',  // 18px
    xl: '1.25rem',   // 20px
    '2xl': '1.5rem', // 24px
    '3xl': '1.875rem', // 30px
  },
} as const