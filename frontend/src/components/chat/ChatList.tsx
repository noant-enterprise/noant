import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn, timeAgo } from '@/lib/utils'
import { Search, MessageCircle, Instagram, Facebook, Send, Globe, Loader2, Trash2 } from 'lucide-react'
import { Avatar } from '@/components/ui/Avatar'
import { useInfiniteScroll } from '@/hooks/useInfiniteScroll'
import type { Conversation } from '@/types'

interface ChatListProps {
  conversations: Conversation[]
  activeId?: string
  hasMore?: boolean
  loadingMore?: boolean
  onLoadMore?: () => void
  onClearAll?: () => void
}

const channelIcons: Record<string, { icon: React.ElementType; color: string }> = {
  whatsapp: { icon: MessageCircle, color: '#25D366' },
  instagram: { icon: Instagram, color: '#E4405F' },
  facebook: { icon: Facebook, color: '#1877F2' },
  telegram: { icon: Send, color: '#0088CC' },
  discord: { icon: MessageCircle, color: '#5865F2' },
  web: { icon: Globe, color: '#0ea5e9' },
}

export function ChatList({
  conversations,
  activeId,
  hasMore = false,
  loadingMore = false,
  onLoadMore,
  onClearAll,
}: ChatListProps) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<'all' | 'active' | 'escalated' | 'resolved'>('all')
  const navigate = useNavigate()

  const filtered = conversations.filter((c) => {
    if (filter !== 'all' && c.status !== filter) return false
    if (search && !c.customer_name.toLowerCase().includes(search.toLowerCase())) return false
    return true
  })

  const { setSentinel } = useInfiniteScroll({
    onLoadMore: onLoadMore || (() => {}),
    hasMore,
    loading: loadingMore,
  })

  return (
    <div className="flex flex-col flex-1 min-h-0 bg-surface">
      {/* Mobile header for chat list */}
      <div className="lg:hidden sticky top-0 z-10 bg-surface/95 backdrop-blur-sm border-b border-default px-4 py-3">
        <h2 className="text-lg font-bold text-primary">Messages</h2>
      </div>

      <div className="p-3 border-b border-default">
        <div className="flex items-center gap-2">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-tertiary" />
            <input
              type="text"
              placeholder="Search conversations..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-8 pr-3 py-2.5 bg-inset border border-default rounded-xl text-sm focus:outline-none focus:border-noant-sky transition-colors"
            />
          </div>
          {onClearAll && conversations.length > 0 && (
            <button
              onClick={onClearAll}
              title="Clear all chats"
              className="p-2.5 rounded-xl border border-default bg-surface hover:bg-red-500/10 text-tertiary hover:text-red-500 hover:border-red-500/20 transition-all shrink-0 active:scale-95 cursor-pointer shadow-sm"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
        <div className="flex gap-1.5 mt-2 overflow-x-auto scrollbar-hide">
          {(['all', 'active', 'escalated', 'resolved'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={cn(
                'px-3 py-1 rounded-full text-[11px] font-semibold capitalize transition-colors shrink-0',
                filter === f ? 'bg-noant-sky text-white' : 'bg-inset text-tertiary hover:text-primary'
              )}
            >
              {f}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto overscroll-contain">
        {filtered.length === 0 && !loadingMore ? (
          <div className="p-8 text-center">
            <div className="w-12 h-12 bg-inset rounded-full flex items-center justify-center mx-auto mb-3">
              <Search className="w-5 h-5 text-tertiary" />
            </div>
            <p className="text-sm font-medium text-secondary">No conversations found</p>
            <p className="text-xs text-tertiary mt-1">Try a different search or filter</p>
          </div>
        ) : (
          <>
            {filtered.map((c) => {
              const isActive = c.id === activeId
              const ch = channelIcons[c.channel] ?? channelIcons.web
              if (!ch) return null
              const Icon = ch.icon
              return (
                <button
                  key={c.id}
                  onClick={() => navigate(`/chats?id=${c.id}`)}
                  className={cn(
                    'w-full flex items-center gap-3 px-4 py-3.5 text-left transition-all duration-200 active:bg-surface-hover',
                    isActive ? 'bg-noant-sky/5' : 'hover:bg-surface-hover',
                    'border-b border-subtle lg:border-default'
                  )}
                >
                  <Avatar
                    src={c.customer_avatar || undefined}
                    name={c.customer_name}
                    size="md"
                    showChannel
                    channelColor={ch.color}
                    channelIcon={<Icon className="w-2 h-2" strokeWidth={3} />}
                  />
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between items-baseline mb-0.5">
                      <span className="font-semibold text-sm text-primary truncate">{c.customer_name}</span>
                      <span className="text-[11px] text-tertiary shrink-0 ml-2">{timeAgo(c.updated_at)}</span>
                    </div>
                    <p className="text-[13px] text-secondary truncate leading-snug">{c.last_message || 'No messages yet'}</p>
                  </div>
                  {c.unread > 0 && (
                    <span className="shrink-0 w-5 h-5 bg-noant-sky text-white text-[10px] font-bold rounded-full flex items-center justify-center">
                      {c.unread > 9 ? '9+' : c.unread}
                    </span>
                  )}
                </button>
              )
            })}
            {hasMore && (
              <div ref={setSentinel} className="py-4 flex justify-center">
                {loadingMore ? (
                  <Loader2 className="w-5 h-5 text-noant-sky animate-spin" />
                ) : (
                  <div className="h-4" />
                )}
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
