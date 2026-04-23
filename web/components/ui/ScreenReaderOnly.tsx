'use client'

import { cn } from '@/lib/utils'
import { ReactNode, ElementType } from 'react'

interface ScreenReaderOnlyProps {
  children: ReactNode
  className?: string
  as?: ElementType
}

export function ScreenReaderOnly({
  children,
  className,
  as: Component = 'span'
}: ScreenReaderOnlyProps) {
  return (
    <Component
      className={cn(
        'absolute w-px h-px p-0 -m-px overflow-hidden',
        'whitespace-nowrap border-0',
        // Alternative: 'sr-only' if using Tailwind
        className
      )}
    >
      {children}
    </Component>
  )
}