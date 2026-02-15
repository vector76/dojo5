interface BalanceSummaryProps {
  expected_balance: number
  total_paid: number
  balance: number
}

export function BalanceSummary({ expected_balance, total_paid, balance }: BalanceSummaryProps) {
  let statusLabel: string
  let statusClass: string
  if (balance === 0) {
    statusLabel = 'Paid in full'
    statusClass = 'balance-paid'
  } else if (balance < 0) {
    statusLabel = 'Overpaid'
    statusClass = 'balance-overpaid'
  } else {
    statusLabel = 'Outstanding'
    statusClass = 'balance-outstanding'
  }

  return (
    <div className={`balance-summary ${statusClass}`} aria-label="Balance Summary">
      <dl>
        <dt>Expected Balance</dt>
        <dd>${expected_balance.toFixed(2)}</dd>
        <dt>Total Paid</dt>
        <dd>${total_paid.toFixed(2)}</dd>
        <dt>Outstanding Balance</dt>
        <dd>${balance.toFixed(2)}</dd>
      </dl>
      <div className="balance-status" data-testid="balance-status">{statusLabel}</div>
    </div>
  )
}
