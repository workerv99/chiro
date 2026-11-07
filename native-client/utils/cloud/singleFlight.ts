/**
 * Wraps an async function so concurrent calls share one in-flight promise.
 * The slot is released once it settles, so the next call starts a new run.
 */
export function singleFlight<T>(fn: () => Promise<T>): () => Promise<T> {
  let inflight: Promise<T> | null = null;
  return () => {
    if (!inflight) {
      inflight = fn().finally(() => { inflight = null; });
    }
    return inflight;
  };
}
