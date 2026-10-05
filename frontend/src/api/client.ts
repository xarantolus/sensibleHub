import createClient from 'openapi-fetch'

import type { ErrorDetail, paths, Problem } from './schema'

export type ApiErrorCode = Problem['code'] | 'network'

export class ApiError extends Error {
  readonly status: number
  readonly code: ApiErrorCode
  readonly problem: Problem | undefined

  constructor(status: number, code: ApiErrorCode, message: string, problem?: Problem) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.problem = problem
  }

  get details(): readonly ErrorDetail[] {
    return this.problem?.errors ?? []
  }

  /** Validation messages keyed by the field name, e.g. `body.title` → `title`. */
  get fieldErrors(): Readonly<Record<string, string>> {
    const out: Record<string, string> = {}
    for (const d of this.details) {
      const field = d.location?.replace(/^body\./, '')
      if (field !== undefined && d.message !== undefined) {
        out[field] = d.message
      }
    }
    return out
  }

  static fromResponse(response: Response, body: unknown): ApiError {
    if (isProblem(body)) {
      return new ApiError(response.status, body.code, body.detail ?? body.title ?? response.statusText, body)
    }
    const code: ApiErrorCode = response.status >= 500 ? 'internal' : 'other'
    return new ApiError(response.status, code, `${String(response.status)} ${response.statusText}`)
  }

  static network(cause: unknown): ApiError {
    const err = new ApiError(0, 'network', 'The server cannot be reached')
    err.cause = cause
    return err
  }
}

function isProblem(body: unknown): body is Problem {
  return typeof body === 'object' && body !== null && 'code' in body && typeof body.code === 'string'
}

export function isApiError(err: unknown, code?: ApiErrorCode): err is ApiError {
  return err instanceof ApiError && (code === undefined || err.code === code)
}

export const client = createClient<paths>({ baseUrl: '/' })

interface FetchResult {
  data?: unknown
  error?: unknown
  response: Response
}

/** Awaits an openapi-fetch call and returns its data, throwing ApiError on any failure. */
export async function call<R extends FetchResult>(request: Promise<R>): Promise<NonNullable<R['data']>> {
  let result: R
  try {
    result = await request
  } catch (err) {
    throw ApiError.network(err)
  }
  if (!result.response.ok) {
    throw ApiError.fromResponse(result.response, result.error)
  }
  return result.data as NonNullable<R['data']>
}

/** Like call, for operations that answer 204 No Content. */
export async function callVoid(request: Promise<FetchResult>): Promise<void> {
  await call(request)
}

export function formData(fields: Record<string, Blob | string>): FormData {
  const fd = new FormData()
  for (const [k, v] of Object.entries(fields)) {
    fd.append(k, v)
  }
  return fd
}
