import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Card, CardContent } from '../card'

describe('Card', () => {
  it('renders its children', () => {
    render(
      <Card>
        <span>card body</span>
      </Card>,
    )
    expect(screen.getByText('card body')).toBeInTheDocument()
  })
})

describe('CardContent', () => {
  it('renders its children', () => {
    render(<CardContent>inner</CardContent>)
    expect(screen.getByText('inner')).toBeInTheDocument()
  })

  it('appends a custom className alongside the base padding', () => {
    render(<CardContent className="custom-class">x</CardContent>)
    const el = screen.getByText('x')
    expect(el).toHaveClass('p-4')
    expect(el).toHaveClass('custom-class')
  })
})
