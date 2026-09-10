import { useEffect, useRef } from "react";

const INITIAL_RETRY_MS = 1000;
const MAX_RETRY_MS = 30000;

/**
 * Subscribes to a Server-Sent Events stream with a bearer token.
 *
 * Uses fetch rather than EventSource on purpose: EventSource cannot set
 * headers, which would force the token into the URL where proxies log it.
 * Reconnects with capped exponential backoff, and reopens whenever the token
 * changes so a refreshed session keeps streaming.
 */
export function useEventStream(
  url: string,
  token: string | null,
  onMessage: (data: string) => void
) {
  // Held in a ref so a new callback identity does not tear down the stream.
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  useEffect(() => {
    if (!token) return;

    const controller = new AbortController();
    let retryMs = INITIAL_RETRY_MS;
    let retryTimer: ReturnType<typeof setTimeout>;

    const connect = async () => {
      try {
        const response = await fetch(url, {
          headers: { Authorization: `Bearer ${token}` },
          signal: controller.signal,
        });

        // A rejected token will not start working on retry.
        if (response.status === 401) return;
        if (!response.ok || !response.body) {
          throw new Error(`stream failed: ${response.status}`);
        }

        retryMs = INITIAL_RETRY_MS;
        await readEvents(response.body, (data) => onMessageRef.current(data));
      } catch {
        if (controller.signal.aborted) return;
      }

      // Stream ended or failed without being aborted: try again later.
      if (!controller.signal.aborted) {
        retryTimer = setTimeout(connect, retryMs);
        retryMs = Math.min(retryMs * 2, MAX_RETRY_MS);
      }
    };

    connect();

    return () => {
      controller.abort();
      clearTimeout(retryTimer);
    };
  }, [url, token]);
}

/**
 * Parses an SSE body. Events are separated by a blank line; each "data:" line
 * contributes one line of the event's payload, and lines starting with ":" are
 * comments (keepalives). Chunks can split an event anywhere, so text is
 * buffered until a complete event arrives.
 */
async function readEvents(
  body: ReadableStream<Uint8Array>,
  onEvent: (data: string) => void
) {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";

  while (true) {
    const { value, done } = await reader.read();
    if (done) return;

    buffer += decoder.decode(value, { stream: true });

    let boundary = buffer.indexOf("\n\n");
    while (boundary !== -1) {
      const rawEvent = buffer.slice(0, boundary);
      buffer = buffer.slice(boundary + 2);

      const data = rawEvent
        .split("\n")
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.slice(5).replace(/^ /, ""))
        .join("\n");

      if (data) onEvent(data);
      boundary = buffer.indexOf("\n\n");
    }
  }
}
