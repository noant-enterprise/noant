import { cn } from '@/lib/utils'

interface ConversationLoadingProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizes = {
  sm: { circle: 60, dots: 40, dot1: 5, dot2: 7, dot3: 9, gap: 4 },
  md: { circle: 80, dots: 56, dot1: 7, dot2: 10, dot3: 13, gap: 6 },
  lg: { circle: 120, dots: 84, dot1: 10, dot2: 14, dot3: 18, gap: 8 },
}

const dotKeyframes = `
  @keyframes dotPulse {
    0%, 100% { transform: scale(0.85); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 1; }
  }
`

export function ConversationLoading({ className, size = 'md' }: ConversationLoadingProps) {
  const s = sizes[size]

  return (
    <>
      <style>{dotKeyframes}</style>
      <div
        className={cn('flex items-center justify-center', className)}
        style={{ width: s.circle, height: s.circle }}
      >
        {/* Inner circle — adapts to theme via CSS variables */}
        <div
          className="rounded-full flex items-center justify-center"
          style={{
            width: s.dots,
            height: s.dots,
            background: 'var(--text-primary)',
            boxShadow: '0 0 25px rgba(0, 0, 0, 0.3)',
          }}
        >
          {/* Pulsing dots — adapts to theme via CSS variables */}
          <div className="flex items-center" style={{ gap: s.gap }}>
            <span
              className="rounded-full"
              style={{
                width: s.dot1,
                height: s.dot1,
                background: 'var(--bg-surface)',
                animation: 'dotPulse 1.4s ease-in-out infinite',
                animationDelay: '0s',
              }}
            />
            <span
              className="rounded-full"
              style={{
                width: s.dot2,
                height: s.dot2,
                background: 'var(--bg-surface)',
                animation: 'dotPulse 1.4s ease-in-out infinite',
                animationDelay: '0.2s',
              }}
            />
            <span
              className="rounded-full"
              style={{
                width: s.dot3,
                height: s.dot3,
                background: 'var(--bg-surface)',
                animation: 'dotPulse 1.4s ease-in-out infinite',
                animationDelay: '0.4s',
              }}
            />
          </div>
        </div>
      </div>
    </>
  )
}
