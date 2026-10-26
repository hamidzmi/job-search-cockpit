export class ApiError extends Error {
  readonly code: string;
  readonly status: number;

  constructor(message: string, code = 'api_error', status = 500) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
    this.status = status;
  }
}

export async function requestJSON<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  });

  if (!response.ok) {
    let message = `Request failed (${response.status})`;
    let code = 'api_error';
    try {
      const payload = await response.json();
      if (payload?.error?.message) {
        message = payload.error.message;
        code = payload.error.code ?? code;
      }
    } catch {
      // ignore non-json error bodies
    }
    throw new ApiError(message, code, response.status);
  }

  return (await response.json()) as T;
}
