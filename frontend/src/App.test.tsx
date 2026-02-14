import { render, screen, waitFor } from '@testing-library/react'
import App from './App'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('App', () => {
  it('renders without errors', async () => {
    render(<App />)
    // Unauthenticated user should be redirected to login
    await waitFor(() => {
      expect(screen.getByText('Login')).toBeInTheDocument()
    })
  })
})
