import { render, screen, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ToastProvider, useToast } from './Toast'

beforeEach(() => {
  vi.restoreAllMocks()
})

function TestTrigger() {
  const { showToast } = useToast()
  return (
    <div>
      <button onClick={() => showToast('Error occurred', 'error')}>Show Error</button>
      <button onClick={() => showToast('Action succeeded', 'success')}>Show Success</button>
    </div>
  )
}

describe('Toast', () => {
  it('renders children without toasts initially', () => {
    render(
      <ToastProvider>
        <div>App content</div>
      </ToastProvider>,
    )

    expect(screen.getByText('App content')).toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('shows a toast when showToast is called', async () => {
    const user = userEvent.setup()

    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    )

    await user.click(screen.getByText('Show Error'))

    expect(screen.getByRole('status')).toBeInTheDocument()
    expect(screen.getByText('Error occurred')).toBeInTheDocument()
  })

  it('auto-dismisses toast after 5 seconds', async () => {
    vi.useFakeTimers()
    // Render and trigger toast using act instead of userEvent (fake timers conflict)
    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    )

    await act(async () => {
      screen.getByText('Show Error').click()
    })
    expect(screen.getByText('Error occurred')).toBeInTheDocument()

    act(() => {
      vi.advanceTimersByTime(5000)
    })

    expect(screen.queryByText('Error occurred')).not.toBeInTheDocument()
    vi.useRealTimers()
  })

  it('dismisses toast when dismiss button is clicked', async () => {
    const user = userEvent.setup()

    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    )

    await user.click(screen.getByText('Show Error'))
    expect(screen.getByText('Error occurred')).toBeInTheDocument()

    await user.click(screen.getByLabelText('Dismiss'))
    expect(screen.queryByText('Error occurred')).not.toBeInTheDocument()
  })

  it('applies correct CSS class for toast type', async () => {
    const user = userEvent.setup()

    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    )

    await user.click(screen.getByText('Show Success'))

    const toast = screen.getByRole('status')
    expect(toast.className).toContain('toast-success')
  })

  it('throws error when useToast is used outside ToastProvider', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    expect(() => render(<TestTrigger />)).toThrow(
      'useToast must be used within a ToastProvider',
    )

    spy.mockRestore()
  })
})
