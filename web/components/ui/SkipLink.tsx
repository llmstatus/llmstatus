'use client'

import { cn } from '@/lib/utils'

interface SkipLinkProps {
  href: string
  children: React.ReactNode
  className?: string
}

export function SkipLink({ href, children, className }: SkipLinkProps) {
  return (
    <a
      href={href}
      className={cn(
        'absolute left-0 top-0 z-50',
        'bg-blue-600 text-white px-4 py-2 rounded',
        'transform -translate-y-full',
        'focus:translate-y-0',
        'transition-transform duration-200',
        className
      )}
    >
      {children}
    </a>
  )
}