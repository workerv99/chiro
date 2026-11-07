/** Error raised for any non-2xx response, timeout or network failure. */
export class CloudApiError extends Error {
  /** HTTP status; 0 for network failures, -1 for timeouts. */
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'CloudApiError';
    this.status = status;
  }

  get isUnauthorized() { return this.status === 401; }
  /** Backend answers 402 when the account has no active Pro plan. */
  get isPaymentRequired() { return this.status === 402; }
  get isNetwork() { return this.status === 0 || this.status === -1; }
}

export const NETWORK_ERROR_MESSAGE = 'No se pudo conectar con el servidor';
export const TIMEOUT_ERROR_MESSAGE = 'El servidor tardó demasiado en responder';

/** Builds a CloudApiError from a status and parsed body (`{ "error": "..." }`). */
export function errorFromResponse(status: number, body: unknown): CloudApiError {
  let message = '';
  if (body && typeof body === 'object') {
    const b = body as Record<string, unknown>;
    if (typeof b.error === 'string') message = b.error;
    else if (typeof b.message === 'string') message = b.message;
  }
  if (!message) message = `Error ${status}`;
  return new CloudApiError(status, message);
}
