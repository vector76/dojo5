import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { ErrorBoundary } from './ErrorBoundary'

beforeEach(() => {
  vi.restoreAllMocks()
})

function ThrowingComponent({ message }: { message: string }) {
  throw new Error(message)
}

describe('ErrorBoundary', () => {
  it('renders children when no error occurs', () => {
    render(
      <ErrorBoundary>
        <div>Safe content</div>
      </ErrorBoundary>,
    )

    expect(screen.getByText('Safe content')).toBeInTheDocument()
  })

  it('renders error UI when a child throws', () => {
    // Suppress React error boundary console output
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    render(
      <ErrorBoundary>
        <ThrowingComponent message="Test error" />
      </ErrorBoundary>,
    )

    expect(screen.getByRole('alert')).toBeInTheDocument()
    expect(screen.getByText('Something went wrong')).toBeInTheDocument()
    expect(screen.getByText('Test error')).toBeInTheDocument()
    expect(screen.getByText('Reload Page')).toBeInTheDocument()

    spy.mockRestore()
  })

  it('shows reload button that reloads the page', async () => {
    const user = userEvent.setup()
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})

    // Mock window.location.reload
    const reloadMock = vi.fn()
    Object.defineProperty(window, 'location', {
      writable: true,
      value: { ...window.location, reload: reloadMock },
    })

    render(
      <ErrorBoundary>
        <ThrowingComponent message="Crash" />
      </ErrorBoundary>,
    )

    await user.click(screen.getByText('Reload Page'))
    expect(reloadMock).toHaveBeenCalled()

    spy.mockRestore()
  })
})
