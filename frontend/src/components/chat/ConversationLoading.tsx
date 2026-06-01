import { cn } from '@/lib/utils'

interface ConversationLoadingProps {
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

const sizes = {
  sm: { circle: 20, outerRing: 15, innerCircle: 11, dot1: 1, dot2: 1.5, dot3: 2, gap: 0.5, strokeWidth: 1, strokeDash: "1 1.5" },
  md: { circle: 60, outerRing: 46, innerCircle: 32, dot1: 3, dot2: 4, dot3: 5, gap: 2, strokeWidth: 3, strokeDash: "3 4" },
  lg: { circle: 120, outerRing: 90, innerCircle: 64, dot1: 5.5, dot2: 8, dot3: 10.5, gap: 4, strokeWidth: 4.5, strokeDash: "5.5 8" },
}

const loaderKeyframes = `
  @keyframes rotateLogoRing {
    from { transform: translate(-50%, -50%) rotate(0deg); }
    to { transform: translate(-50%, -50%) rotate(360deg); }
  }

  @keyframes logoDotPulse {
    0%, 100% { transform: scale(0.85); opacity: 0.6; }
    50% { transform: scale(1.1); opacity: 1; }
  }
`

export function ConversationLoading({ className, size = 'md' }: ConversationLoadingProps) {
  const s = sizes[size]

  return (
    <>
      <style>{loaderKeyframes}</style>
      <div
        className={cn('relative flex items-center justify-center shrink-0', className)}
        style={{
          width: s.circle,
          height: s.circle,
        }}
      >
        {/* Spinning dashed ring */}
        <svg
          className="absolute top-1/2 left-1/2"
          style={{
            width: s.outerRing,
            height: s.outerRing,
            transform: 'translate(-50%, -50%)',
            animation: 'rotateLogoRing 8s linear infinite',
          }}
          viewBox="0 0 148 148"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
        >
          <circle
            cx="74"
            cy="74"
            r="68"
            stroke="var(--text-primary)"
            strokeWidth={7}
            strokeDasharray="9 13"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>

        {/* Inner circle with pulsing dots */}
        <div
          className="absolute top-1/2 left-1/2 rounded-full flex items-center justify-center"
          style={{
            width: s.innerCircle,
            height: s.innerCircle,
            background: 'var(--text-primary)',
            border: `${size === 'sm' ? 1 : size === 'md' ? 3 : 6}px solid var(--text-primary)`,
            transform: 'translate(-50%, -50%)',
          }}
        >
          <div className="flex items-center" style={{ gap: s.gap }}>
            <div
              className="rounded-full"
              style={{
                width: s.dot1,
                height: s.dot1,
                background: 'var(--bg-surface)',
                animation: 'logoDotPulse 1.4s ease-in-out infinite',
                animationDelay: '0s',
              }}
            />
            <div
              className="rounded-full"
              style={{
                width: s.dot2,
                height: s.dot2,
                background: 'var(--bg-surface)',
                animation: 'logoDotPulse 1.4s ease-in-out infinite',
                animationDelay: '0.2s',
              }}
            />
            <div
              className="rounded-full"
              style={{
                width: s.dot3,
                height: s.dot3,
                background: 'var(--bg-surface)',
                animation: 'logoDotPulse 1.4s ease-in-out infinite',
                animationDelay: '0.4s',
              }}
            />
          </div>
        </div>
      </div>
    </>
  )
}
