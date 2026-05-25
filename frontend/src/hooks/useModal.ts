import { useState, useCallback } from 'react'

interface UseModalReturn {
  open: boolean
  openModal: () => void
  closeModal: () => void
  toggle: () => void
}

export function useModal(initial = false): UseModalReturn {
  const [open, setOpen] = useState(initial)

  const openModal = useCallback(() => setOpen(true), [])
  const closeModal = useCallback(() => setOpen(false), [])
  const toggle = useCallback(() => setOpen((p) => !p), [])

  return { open, openModal, closeModal, toggle }
}
