import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QAPair } from '@/components/training/QAPair'

const mockQA = {
  id: 'qa-1',
  question: 'How do I reset my password?',
  answer: 'Go to Settings > Security and click Reset Password.',
}

describe('QAPair', () => {
  it('renders the question', () => {
    render(<QAPair qa={mockQA} />)
    expect(screen.getByText('How do I reset my password?')).toBeInTheDocument()
  })

  it('renders the answer', () => {
    render(<QAPair qa={mockQA} />)
    expect(screen.getByText('Go to Settings > Security and click Reset Password.')).toBeInTheDocument()
  })

  it('renders a border on the answer', () => {
    render(<QAPair qa={mockQA} />)
    const answer = screen.getByText('Go to Settings > Security and click Reset Password.')
    expect(answer.className).toContain('border-l-2')
  })

  it('renders a different QA pair', () => {
    const otherQA = {
      id: 'qa-2',
      question: 'What are your hours?',
      answer: 'We are open 9-5.',
    }
    render(<QAPair qa={otherQA} />)
    expect(screen.getByText('What are your hours?')).toBeInTheDocument()
    expect(screen.getByText('We are open 9-5.')).toBeInTheDocument()
  })
})
