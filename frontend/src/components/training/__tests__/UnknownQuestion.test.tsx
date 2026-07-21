import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { UnknownQuestionItem } from '@/components/training/UnknownQuestion'

const mockQuestion = {
  id: 'uq-1',
  question: 'What is the return policy?',
}

const mockOnTrain = vi.fn()
const mockOnIgnore = vi.fn()

beforeEach(() => {
  vi.clearAllMocks()
  mockOnTrain.mockResolvedValue(undefined)
})

describe('UnknownQuestionItem', () => {
  it('renders the question text', () => {
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    expect(screen.getByText('"What is the return policy?"')).toBeInTheDocument()
  })

  it('renders Train button', () => {
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    expect(screen.getByRole('button', { name: /train/i })).toBeInTheDocument()
  })

  it('renders Ignore button', () => {
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    expect(screen.getByRole('button', { name: /ignore/i })).toBeInTheDocument()
  })

  it('calls onIgnore with question id when Ignore is clicked', async () => {
    const user = userEvent.setup()
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    await user.click(screen.getByRole('button', { name: /ignore/i }))
    expect(mockOnIgnore).toHaveBeenCalledWith('uq-1')
  })

  it('shows training form when Train is clicked', async () => {
    const user = userEvent.setup()
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    await user.click(screen.getByRole('button', { name: /train/i }))
    expect(screen.getByText('Train AI Response')).toBeInTheDocument()
    expect(screen.getByLabelText('Correct Answer')).toBeInTheDocument()
    expect(screen.getByLabelText('Select Category')).toBeInTheDocument()
  })

  it('submits training form with answer and category', async () => {
    const user = userEvent.setup()
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    await user.click(screen.getByRole('button', { name: /train/i }))

    await user.type(screen.getByLabelText('Correct Answer'), '30-day return policy')
    await user.click(screen.getByRole('button', { name: /save answer/i }))

    expect(mockOnTrain).toHaveBeenCalledWith('uq-1', '30-day return policy', 'default')
  })

  it('does not submit when answer is empty', async () => {
    const user = userEvent.setup()
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    await user.click(screen.getByRole('button', { name: /train/i }))
    await user.click(screen.getByRole('button', { name: /save answer/i }))

    expect(mockOnTrain).not.toHaveBeenCalled()
  })

  it('cancels training form', async () => {
    const user = userEvent.setup()
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
      />
    )
    await user.click(screen.getByRole('button', { name: /train/i }))
    expect(screen.getByText('Train AI Response')).toBeInTheDocument()

    const cancelButtons = screen.getAllByRole('button', { name: /cancel/i })
    await user.click(cancelButtons[0]!)

    expect(screen.queryByText('Train AI Response')).not.toBeInTheDocument()
  })

  it('renders categories in select dropdown', async () => {
    const user = userEvent.setup()
    const categories = [
      { id: 'default', name: 'Default' },
      { id: 'billing', name: 'Billing' },
      { id: 'support', name: 'Support' },
    ]
    render(
      <UnknownQuestionItem
        question={mockQuestion}
        onTrain={mockOnTrain}
        onIgnore={mockOnIgnore}
        categories={categories}
      />
    )
    await user.click(screen.getByRole('button', { name: /train/i }))
    expect(screen.getByText('Billing')).toBeInTheDocument()
    expect(screen.getByText('Support')).toBeInTheDocument()
  })
})
