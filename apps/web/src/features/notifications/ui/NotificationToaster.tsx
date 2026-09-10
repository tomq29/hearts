import { useCallback, useState } from "react";
import { useAuthStore } from "@/entities/session";
import { useEventStream } from "@/shared/lib/hooks/useEventStream";
import { HeartIcon } from "@/shared/ui/icons";

const NOTIFICATIONS_URL =
  (import.meta.env.VITE_NOTIFICATIONS_URL ?? "http://localhost:8081") +
  "/notifications/stream";

const DISMISS_AFTER_MS = 6000;

interface Toast {
  id: number;
  type: string;
  message: string;
}

let nextId = 0;

/**
 * Listens to the live notification stream and shows each notification as a
 * toast. Mounted once at the app root, for signed-in users only.
 */
export const NotificationToaster = () => {
  const token = useAuthStore((state) => state.token);
  const [toasts, setToasts] = useState<Toast[]>([]);

  const dismiss = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id));
  }, []);

  const handleMessage = useCallback(
    (data: string) => {
      let payload: { type?: string; message?: string };
      try {
        payload = JSON.parse(data);
      } catch {
        return; // Ignore anything that is not a notification.
      }
      if (!payload.message) return;

      const id = nextId++;
      setToasts((current) => [
        ...current,
        { id, type: payload.type ?? "info", message: payload.message! },
      ]);
      setTimeout(() => dismiss(id), DISMISS_AFTER_MS);
    },
    [dismiss]
  );

  useEventStream(NOTIFICATIONS_URL, token, handleMessage);

  if (toasts.length === 0) return null;

  return (
    <div
      className="fixed bottom-4 right-4 z-50 flex flex-col gap-2 w-80 max-w-[calc(100vw-2rem)]"
      aria-live="polite"
    >
      {toasts.map((toast) => (
        <div
          key={toast.id}
          role="status"
          className="flex items-start gap-3 bg-white border border-gray-100 shadow-lg rounded-xl p-4"
        >
          {toast.type === "match" && (
            <HeartIcon className="w-5 h-5 text-pink-600 shrink-0 mt-0.5" />
          )}
          <p className="text-sm text-gray-800 flex-1">{toast.message}</p>
          <button
            type="button"
            onClick={() => dismiss(toast.id)}
            className="text-gray-400 hover:text-gray-600 text-lg leading-none"
            aria-label="Dismiss notification"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
};
