import { useState, useMemo } from 'react'
import { cn } from '@/lib/utils'

interface AvatarProps {
  src?: string
  name: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
  showChannel?: boolean
  channelColor?: string
  channelIcon?: React.ReactNode
}

const SIZE_MAP = {
  sm: 'w-7 h-7 text-[10px]',
  md: 'w-8 h-8 text-xs',
  lg: 'w-10 h-10 text-sm',
  xl: 'w-12 h-12 text-base',
}

const GRADIENTS = [
  'from-slate-500 to-slate-700',
  'from-sky-500 to-blue-600',
  'from-emerald-500 to-teal-600',
  'from-amber-500 to-orange-600',
  'from-rose-500 to-pink-600',
  'from-violet-500 to-purple-600',
  'from-cyan-500 to-blue-500',
  'from-lime-500 to-green-600',
]

function hashName(name: string): number {
  let hash = 0
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash)
  }
  return Math.abs(hash)
}

export function Avatar({
  src,
  name,
  size = 'md',
  className,
  showChannel,
  channelColor,
  channelIcon,
}: AvatarProps) {
  const [imgError, setImgError] = useState(false)
  const initials = useMemo(() => {
    return name
      .split(' ')
      .map((n) => n[0])
      .join('')
      .toUpperCase()
      .slice(0, 2)
  }, [name])

  const gradient = useMemo(() => {
    return GRADIENTS[hashName(name) % GRADIENTS.length]
  }, [name])

  const showImage = src && !imgError

  return (
    <div className={cn('relative shrink-0 inline-flex', className)}>
      {showImage ? (
        <img
          src={src}
          alt={name}
          onError={() => setImgError(true)}
          className={cn(
            'rounded-full object-cover border border-default',
            SIZE_MAP[size]
          )}
        />
      ) : (
        <div
          className={cn(
            'rounded-full flex items-center justify-center font-bold text-white bg-gradient-to-br',
            gradient,
            SIZE_MAP[size]
          )}
        >
          {initials}
        </div>
      )}
      {showChannel && channelColor && (
        <div
          className="absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full flex items-center justify-center border-2 border-surface"
          style={{ background: channelColor }}
        >
          {channelIcon && (
            <span className="text-white scale-75">{channelIcon}</span>
          )}
        </div>
      )}
    </div>
  )
}