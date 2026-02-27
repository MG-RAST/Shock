/** Error returned when the Shock server responds with a non-2xx status. */
export class ShockError extends Error {
  readonly status: number;
  readonly messages: string[];

  constructor(status: number, messages: string[]) {
    super(messages.join("; ") || `Shock error (HTTP ${status})`);
    this.name = "ShockError";
    this.status = status;
    this.messages = messages;
  }

  static fromResponse(status: number, error: string[] | null): ShockError {
    const msgs = error ?? [`HTTP ${status}`];
    if (status === 423) {
      return new ShockLockedError(status, msgs);
    }
    return new ShockError(status, msgs);
  }
}

/** Network-level error (fetch failed, timeout, DNS, etc.). */
export class ShockNetworkError extends Error {
  readonly networkCause?: unknown;

  constructor(message: string, cause?: unknown) {
    super(message);
    this.name = "ShockNetworkError";
    this.networkCause = cause;
  }
}

/**
 * HTTP 423 — the node's file is locked (being assembled or indexed).
 * Callers should retry after a delay.
 */
export class ShockLockedError extends ShockError {
  constructor(status: number, messages: string[]) {
    super(status, messages);
    this.name = "ShockLockedError";
  }
}
