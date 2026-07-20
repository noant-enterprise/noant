import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UploadZone } from '@/components/training/UploadZone'

const mockOnUpload = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
})

describe('UploadZone', () => {
  it('renders the drop zone text', () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    expect(screen.getByText('Drop your CSV here')).toBeInTheDocument()
  })

  it('renders the format hint', () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    expect(screen.getByText('Format: Category, Question, Answer')).toBeInTheDocument()
  })

  it('renders the choose file button', () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    expect(screen.getByText('Choose file')).toBeInTheDocument()
  })

  it('renders the hidden file input', () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    const input = document.getElementById('csv-upload') as HTMLInputElement
    expect(input).toBeInTheDocument()
    expect(input.type).toBe('file')
    expect(input.accept).toBe('.csv')
  })

  it('does not show progress when not uploading', () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    expect(screen.queryByText(/uploaded/)).not.toBeInTheDocument()
  })

  it('shows progress when uploading', () => {
    render(<UploadZone onUpload={mockOnUpload} uploading progress={60} />)
    expect(screen.getByText('60% uploaded')).toBeInTheDocument()
  })

  it('calls onUpload when a CSV file is selected via input', async () => {
    const user = userEvent.setup()
    render(<UploadZone onUpload={mockOnUpload} />)

    const file = new File(['name,question,answer'], 'data.csv', { type: 'text/csv' })
    const input = document.getElementById('csv-upload') as HTMLInputElement

    await user.upload(input, file)
    expect(mockOnUpload).toHaveBeenCalledWith(file)
  })

  it('does not upload non-csv files via drop', async () => {
    render(<UploadZone onUpload={mockOnUpload} />)
    const dropZone = screen.getByText('Drop your CSV here').closest('div')!.parentElement!

    const file = new File(['data'], 'data.txt', { type: 'text/plain' })
    const dataTransfer = { files: [file] }
    const dropEvent = new Event('drop', { bubbles: true, cancelable: true }) as any
    dropEvent.dataTransfer = dataTransfer
    vi.spyOn(dropEvent, 'preventDefault')

    dropZone.dispatchEvent(dropEvent)
    expect(mockOnUpload).not.toHaveBeenCalled()
  })
})
