export interface ApiResponse<T = unknown> {
  message: string
  data: T
  success: boolean
}

export class ApiError extends Error {
  status: number
  data: Record<string, string>

  constructor(message: string, status: number, data: Record<string, string> = {}) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) {
    return error
  }

  const err = error as {
    statusCode?: number
    statusMessage?: string
    data?: ApiResponse<Record<string, string>>
  }

  const body = err.data
  if (body && typeof body === 'object' && 'message' in body) {
    return new ApiError(
      body.message || 'Request failed',
      err.statusCode || 500,
      (body.data && typeof body.data === 'object' ? body.data : {}) as Record<string, string>
    )
  }

  return new ApiError(err.statusMessage || 'Request failed', err.statusCode || 500)
}
