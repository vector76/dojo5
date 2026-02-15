import { render, screen } from '@testing-library/react'
import { BalanceSummary } from './BalanceSummary'

describe('BalanceSummary', () => {
  it('renders balance values', () => {
    render(<BalanceSummary expected_balance={100} total_paid={60} balance={40} />)

    expect(screen.getByText('$100.00')).toBeInTheDocument()
    expect(screen.getByText('$60.00')).toBeInTheDocument()
    expect(screen.getByText('$40.00')).toBeInTheDocument()
  })

  it('shows "Outstanding" status when balance is positive', () => {
    render(<BalanceSummary expected_balance={100} total_paid={60} balance={40} />)

    expect(screen.getByTestId('balance-status')).toHaveTextContent('Outstanding')
    expect(screen.getByLabelText('Balance Summary')).toHaveClass('balance-outstanding')
  })

  it('shows "Paid in full" status when balance is zero', () => {
    render(<BalanceSummary expected_balance={100} total_paid={100} balance={0} />)

    expect(screen.getByTestId('balance-status')).toHaveTextContent('Paid in full')
    expect(screen.getByLabelText('Balance Summary')).toHaveClass('balance-paid')
  })

  it('shows "Overpaid" status when balance is negative', () => {
    render(<BalanceSummary expected_balance={100} total_paid={120} balance={-20} />)

    expect(screen.getByTestId('balance-status')).toHaveTextContent('Overpaid')
    expect(screen.getByLabelText('Balance Summary')).toHaveClass('balance-overpaid')
    expect(screen.getByText('$-20.00')).toBeInTheDocument()
  })

  it('handles zero expected balance', () => {
    render(<BalanceSummary expected_balance={0} total_paid={0} balance={0} />)

    const zeroCells = screen.getAllByText('$0.00')
    expect(zeroCells).toHaveLength(3)
    expect(screen.getByTestId('balance-status')).toHaveTextContent('Paid in full')
  })

  it('renders with aria-label for accessibility', () => {
    render(<BalanceSummary expected_balance={50} total_paid={25} balance={25} />)

    expect(screen.getByLabelText('Balance Summary')).toBeInTheDocument()
  })
})
