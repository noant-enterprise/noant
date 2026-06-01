import { useEffect, useState } from 'react'
import { useAPI } from '@/hooks/useAPI'
import { useToast } from '@/components/ui/Toast'
import { Card, CardBody } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { StatCard, StatGrid } from '@/components/stats'
import { StatSkeleton } from '@/components/ui/Skeleton'
import { Package, Plus, Edit2, Trash2, Search, X } from 'lucide-react'
import { api } from '@/lib/api'

interface InventoryItem {
  id: string
  type: 'product' | 'service' | 'package'
  name: string
  description: string
  price: number
  min_price: number | null
  stock_quantity: number | null
  image_url: string | null
  is_active: boolean
  created_at: string
}

const typeConfig: Record<string, { label: string; color: string }> = {
  product: { label: 'Product', color: 'sky' },
  service: { label: 'Service', color: 'success' },
  package: { label: 'Package', color: 'warning' },
}

export default function InventoryPage() {
  const { toast } = useToast()
  const apiHook = useAPI() as any
  const { data, get: getItems, loading } = apiHook
  const [showModal, setShowModal] = useState(false)
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState('')

  const items: InventoryItem[] = data?.items || []

  useEffect(() => {
    loadItems()
  }, [typeFilter])

  const loadItems = () => {
    const params = new URLSearchParams()
    if (typeFilter) params.set('type', typeFilter)
    const query = params.toString()
    getItems(`/inventory${query ? `?${query}` : ''}`)
  }

  const handleSearch = async () => {
    if (!searchQuery.trim()) {
      loadItems()
      return
    }
    try {
      const res = await api.get<any>(`/inventory/search?q=${encodeURIComponent(searchQuery)}`)
      // Update local state with search results
      if (data) data.items = res.items || []
    } catch {
      toast('Search failed', 'error')
    }
  }

  const handleCreate = async (item: Partial<InventoryItem>) => {
    try {
      await api.post('/inventory', item)
      toast('Item created', 'success')
      setShowModal(false)
      loadItems()
    } catch {
      toast('Failed to create item', 'error')
    }
  }

  const handleUpdate = async (item: Partial<InventoryItem>) => {
    try {
      await api.put(`/inventory/${item.id}`, item)
      toast('Item updated', 'success')
      setShowModal(false)
      setEditingItem(null)
      loadItems()
    } catch {
      toast('Failed to update item', 'error')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this item?')) return
    try {
      await api.delete(`/inventory/${id}`)
      toast('Item deleted', 'success')
      loadItems()
    } catch {
      toast('Failed to delete item', 'error')
    }
  }

  const totalItems = items.length
  const totalValue = items.reduce((sum, i) => sum + i.price * (i.stock_quantity || 1), 0)
  const products = items.filter(i => i.type === 'product').length
  const services = items.filter(i => i.type === 'service').length

  return (
    <div className="animate-page-in space-y-5 lg:space-y-6 pt-2">
      {/* Stats */}
      <div className="px-1">
        {loading ? (
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatSkeleton /><StatSkeleton /><StatSkeleton /><StatSkeleton />
          </div>
        ) : (
          <StatGrid>
            <StatCard label="Total Items" value={totalItems} variant="default" />
            <StatCard label="Products" value={products} variant="info" />
            <StatCard label="Services" value={services} variant="success" />
            <StatCard label="Total Value" value={`₦${totalValue.toLocaleString()}`} variant="warning" />
          </StatGrid>
        )}
      </div>

      {/* Header + Search */}
      <div className="px-1">
        <div className="flex flex-col sm:flex-row gap-3 items-start sm:items-center justify-between">
          <div className="flex gap-2 flex-wrap">
            {['', 'product', 'service', 'package'].map(t => (
              <button
                key={t}
                onClick={() => setTypeFilter(t)}
                className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                  typeFilter === t
                    ? 'bg-noant-sky text-white'
                    : 'bg-inset text-secondary hover:text-primary'
                }`}
              >
                {t === '' ? 'All' : typeConfig[t]?.label || t}
              </button>
            ))}
          </div>
          <div className="flex gap-2 w-full sm:w-auto">
            <div className="relative flex-1 sm:w-64">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-tertiary" />
              <input
                type="text"
                placeholder="Search items..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                onKeyDown={e => e.key === 'Enter' && handleSearch()}
                className="w-full pl-9 pr-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30"
              />
            </div>
            <Button onClick={() => { setEditingItem(null); setShowModal(true) }}>
              <Plus className="w-4 h-4 mr-1.5" /> Add Item
            </Button>
          </div>
        </div>
      </div>

      {/* Items grid */}
      <div className="px-1">
        {loading ? (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="animate-shimmer-slow h-40 rounded-xl" />
            ))}
          </div>
        ) : items.length === 0 ? (
          <Card>
            <CardBody className="p-8 text-center">
              <Package className="w-12 h-12 mx-auto text-tertiary mb-3" />
              <p className="text-secondary text-sm">No inventory items yet</p>
              <p className="text-tertiary text-xs mt-1">Add products, services, or packages for the AI to sell</p>
              <Button className="mt-4" onClick={() => { setEditingItem(null); setShowModal(true) }}>
                <Plus className="w-4 h-4 mr-1.5" /> Add First Item
              </Button>
            </CardBody>
          </Card>
        ) : (
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
            {items.map(item => {
              const cfg = typeConfig[item.type]
              return (
                <Card key={item.id} className="hover:shadow-lg transition-shadow">
                  <CardBody className="p-4">
                    <div className="flex items-start justify-between mb-3">
                      <div className="flex items-center gap-2">
                        <div className="w-10 h-10 rounded-xl bg-noant-sky/10 flex items-center justify-center">
                          <Package className="w-5 h-5 text-noant-sky" />
                        </div>
                        <div>
                          <div className="text-sm font-semibold text-primary">{item.name}</div>
                          <Badge variant={cfg?.color as any || 'neutral'} className="text-[10px] mt-0.5">{cfg?.label || item.type}</Badge>
                        </div>
                      </div>
                      <div className="flex gap-1">
                        <button
                          onClick={() => { setEditingItem(item); setShowModal(true) }}
                          className="w-7 h-7 rounded-lg bg-inset flex items-center justify-center text-tertiary hover:text-noant-sky transition-colors"
                        >
                          <Edit2 className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => handleDelete(item.id)}
                          className="w-7 h-7 rounded-lg bg-inset flex items-center justify-center text-tertiary hover:text-red-500 transition-colors"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>
                    {item.description && (
                      <p className="text-xs text-tertiary mb-3 line-clamp-2">{item.description}</p>
                    )}
                    <div className="flex items-end justify-between">
                      <div>
                        <div className="text-lg font-bold text-primary">₦{item.price.toLocaleString()}</div>
                        {item.min_price && item.min_price < item.price && (
                          <div className="text-[10px] text-tertiary">Min: ₦{item.min_price.toLocaleString()}</div>
                        )}
                      </div>
                      {item.stock_quantity !== null && (
                        <div className="text-right">
                          <div className="text-xs text-tertiary">Stock</div>
                          <div className={`text-sm font-medium ${item.stock_quantity <= 5 ? 'text-red-500' : 'text-primary'}`}>
                            {item.stock_quantity}
                          </div>
                        </div>
                      )}
                    </div>
                  </CardBody>
                </Card>
              )
            })}
          </div>
        )}
      </div>

      {/* Create/Edit Modal */}
      {showModal && (
        <InventoryModal
          item={editingItem}
          onClose={() => { setShowModal(false); setEditingItem(null) }}
          onSave={editingItem ? handleUpdate : handleCreate}
        />
      )}
    </div>
  )
}

function InventoryModal({ item, onClose, onSave }: { item: InventoryItem | null; onClose: () => void; onSave: (data: any) => void }) {
  const [form, setForm] = useState({
    type: item?.type || 'product',
    name: item?.name || '',
    description: item?.description || '',
    price: item?.price || 0,
    min_price: item?.min_price || '',
    stock_quantity: item?.stock_quantity ?? '',
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.name || !form.price) return
    onSave({
      ...form,
      id: item?.id,
      price: Number(form.price),
      min_price: form.min_price ? Number(form.min_price) : null,
      stock_quantity: form.stock_quantity !== '' ? Number(form.stock_quantity) : null,
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative bg-surface border border-default rounded-2xl w-full max-w-md p-6 shadow-xl">
        <div className="flex items-center justify-between mb-5">
          <h3 className="text-lg font-semibold text-primary">{item ? 'Edit Item' : 'Add Item'}</h3>
          <button onClick={onClose} className="w-8 h-8 rounded-full bg-inset flex items-center justify-center text-tertiary hover:text-primary">
            <X className="w-4 h-4" />
          </button>
        </div>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-xs font-medium text-secondary mb-1.5">Type</label>
            <div className="flex gap-2">
              {['product', 'service', 'package'].map(t => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setForm(f => ({ ...f, type: t as 'product' | 'service' | 'package' }))}
                  className={`flex-1 py-2 rounded-xl text-xs font-medium transition-all ${
                    form.type === t ? 'bg-noant-sky text-white' : 'bg-inset text-secondary'
                  }`}
                >
                  {typeConfig[t]?.label || t}
                </button>
              ))}
            </div>
          </div>
          <div>
            <label className="block text-xs font-medium text-secondary mb-1.5">Name *</label>
            <input
              type="text"
              value={form.name}
              onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
              className="w-full px-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30"
              placeholder="e.g. Premium Package"
              required
            />
          </div>
          <div>
            <label className="block text-xs font-medium text-secondary mb-1.5">Description</label>
            <textarea
              value={form.description}
              onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
              className="w-full px-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30 resize-none"
              rows={2}
              placeholder="Brief description..."
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-xs font-medium text-secondary mb-1.5">Price (₦) *</label>
              <input
                type="number"
                value={form.price}
                onChange={e => setForm(f => ({ ...f, price: Number(e.target.value) }))}
                className="w-full px-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30"
                min="0"
                required
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-secondary mb-1.5">Min Price (₦)</label>
              <input
                type="number"
                value={form.min_price}
                onChange={e => setForm(f => ({ ...f, min_price: e.target.value }))}
                className="w-full px-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30"
                min="0"
                placeholder="Floor price"
              />
            </div>
          </div>
          {form.type === 'product' && (
            <div>
              <label className="block text-xs font-medium text-secondary mb-1.5">Stock Quantity</label>
              <input
                type="number"
                value={form.stock_quantity}
                onChange={e => setForm(f => ({ ...f, stock_quantity: e.target.value }))}
                className="w-full px-3 py-2 bg-inset border border-default rounded-xl text-sm text-primary placeholder:text-tertiary focus:outline-none focus:ring-2 focus:ring-noant-sky/30"
                min="0"
                placeholder="Units available"
              />
            </div>
          )}
          <div className="flex gap-2 pt-2">
            <Button type="button" variant="ghost" className="flex-1" onClick={onClose}>Cancel</Button>
            <Button type="submit" className="flex-1">{item ? 'Update' : 'Create'}</Button>
          </div>
        </form>
      </div>
    </div>
  )
}
