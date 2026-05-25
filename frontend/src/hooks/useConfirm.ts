import { useCallback } from 'react'
import { useModalContext, type ModalOptions } from '@/contexts/ModalContext'

/**
 * useConfirm — thin wrapper around ModalContext.showModal for quick usage.
 *
 * Usage:
 *   const confirm = useConfirm()
 *
 *   confirm({
 *     title: 'Delete Category?',
 *     body: 'This cannot be undone.',
 *     variant: 'danger',
 *     confirmText: 'Delete',
 *     onConfirm: async () => { await api.delete(...) },
 *   })
 */
export function useConfirm() {
  const { showModal } = useModalContext()

  const confirm = useCallback(
    (options: ModalOptions) => {
      showModal(options)
    },
    [showModal]
  )

  return confirm
}
