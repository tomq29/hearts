/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  /** Backend origin, e.g. https://api.hearts.example.com (dev default: http://localhost:8080) */
  readonly VITE_API_URL?: string;
  /** Notification service origin (dev default: http://localhost:8081) */
  readonly VITE_NOTIFICATIONS_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
