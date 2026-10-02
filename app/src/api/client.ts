import { tokenStore } from './tokenStore';

const apiBaseUrl = (process.env.EXPO_PUBLIC_API_URL ?? 'http://localhost:8080').replace(/\/$/, '');

export type ApiErrorCode =
  | 'validation'
  | 'invalid_credentials'
  | 'unauthenticated'
  | 'account_locked'
  | 'forbidden'
  | 'not_found'
  | 'conflict'
  | 'internal'
  | 'network';

export class ApiError extends Error {
  constructor(
    readonly status: number,
    readonly code: ApiErrorCode,
    message: string,
    readonly field?: string,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE';

export async function apiRequest<T>(method: HttpMethod, path: string, body?: unknown): Promise<T> {
  let response: Response;
  try {
    response = await fetch(`${apiBaseUrl}${path}`, {
      method,
      credentials: 'include',
      headers: await buildHeaders(body !== undefined),
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new ApiError(0, 'network', 'No se pudo conectar con el servidor. Revisá tu conexión.');
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const payload = await response.json().catch(() => null);
  if (!response.ok) {
    throw toApiError(response.status, payload);
  }
  return payload as T;
}

async function buildHeaders(hasBody: boolean): Promise<Record<string, string>> {
  const headers: Record<string, string> = { Accept: 'application/json' };
  if (hasBody) {
    headers['Content-Type'] = 'application/json';
  }
  if (tokenStore.usesCookies) {
    return headers;
  }

  headers['X-Client-Platform'] = 'native';
  const token = await tokenStore.read();
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

function toApiError(status: number, payload: unknown): ApiError {
  const body = (payload ?? {}) as { code?: ApiErrorCode; message?: string; field?: string };
  return new ApiError(status, body.code ?? 'internal', body.message ?? 'Error inesperado del servidor', body.field);
}
