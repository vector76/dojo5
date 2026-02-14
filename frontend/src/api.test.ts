import { apiFetch, ApiError, getToken, setToken, clearToken } from './api'

beforeEach(() => {
  localStorage.clear()
  vi.restoreAllMocks()
})

describe('token helpers', () => {
  it('getToken returns null when no token is stored', () => {
    expect(getToken()).toBeNull()
  })

  it('setToken stores and getToken retrieves a token', () => {
    setToken('my-jwt')
    expect(getToken()).toBe('my-jwt')
  })

  it('clearToken removes the token', () => {
    setToken('my-jwt')
    clearToken()
    expect(getToken()).toBeNull()
  })
})

describe('apiFetch', () => {
  it('attaches Authorization header when token exists', async () => {
    setToken('test-token')
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('{}', { status: 200 }),
    )

    await apiFetch('/api/test')

    expect(fetch).toHaveBeenCalledOnce()
    const [, init] = vi.mocked(fetch).mock.calls[0]
    const headers = init?.headers as Headers
    expect(headers.get('Authorization')).toBe('Bearer test-token')
  })

  it('does not attach Authorization header when no token', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('{}', { status: 200 }),
    )

    await apiFetch('/api/test')

    const [, init] = vi.mocked(fetch).mock.calls[0]
    const headers = init?.headers as Headers
    expect(headers.get('Authorization')).toBeNull()
  })

  it('sets Content-Type to application/json when body is provided', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('{}', { status: 200 }),
    )

    await apiFetch('/api/test', {
      method: 'POST',
      body: JSON.stringify({ key: 'value' }),
    })

    const [, init] = vi.mocked(fetch).mock.calls[0]
    const headers = init?.headers as Headers
    expect(headers.get('Content-Type')).toBe('application/json')
  })

  it('throws ApiError on non-ok response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('not found', { status: 404 }),
    )

    await expect(apiFetch('/api/missing')).rejects.toThrow(ApiError)

    try {
      await apiFetch('/api/missing')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).status).toBe(404)
    }
  })

  it('throws ApiError with status 401 on unauthorized response', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response('unauthorized', { status: 401 }),
    )

    try {
      await apiFetch('/api/protected')
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError)
      expect((err as ApiError).status).toBe(401)
    }
  })

  it('returns response on success', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify({ ok: true }), { status: 200 }),
    )

    const res = await apiFetch('/api/test')
    const data = await res.json()
    expect(data).toEqual({ ok: true })
  })
})
