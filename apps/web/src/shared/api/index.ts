import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "@/entities/session";

// Configured per environment via VITE_API_URL (see .env.example);
// falls back to the local dev backend.
export const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

// WebSocket endpoint derived from the API origin (http -> ws, https -> wss).
export const WS_URL = API_URL.replace(/^http/, "ws") + "/ws";

export const api = axios.create({
  baseURL: API_URL,
  // Required for the HttpOnly refreshToken cookie to be stored and sent
  // on cross-origin requests (frontend :5173 -> API :8080).
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Single-flight refresh: if several requests hit 401 at once, only one
// refresh call is made and the rest await its result.
let refreshPromise: Promise<string> | null = null;

const refreshAccessToken = async (): Promise<string> => {
  // Bare axios (not `api`) so a failed refresh doesn't re-enter the
  // interceptor below and loop forever.
  const response = await axios.post<{ accessToken: string }>(
    `${API_URL}/api/v1/auth/refresh`,
    null,
    { withCredentials: true }
  );
  return response.data.accessToken;
};

// restoreSession recovers the session after a page reload: the access token
// lives in memory only, so we exchange the HttpOnly refresh cookie for a
// fresh one. A failure just means the user isn't logged in.
export const restoreSession = async (): Promise<void> => {
  try {
    const token = await refreshAccessToken();
    useAuthStore.getState().setToken(token);
  } catch {
    // No valid refresh cookie — stay logged out.
  }
};

// 401s from credential endpoints mean "wrong email/password", not
// "access token expired" — refreshing won't help there.
const AUTH_URLS = ["/api/v1/users/login", "/api/v1/users/register"];

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as
      | (InternalAxiosRequestConfig & { _retry?: boolean })
      | undefined;

    const shouldAttemptRefresh =
      error.response?.status === 401 &&
      originalRequest &&
      !originalRequest._retry &&
      !AUTH_URLS.some((url) => originalRequest.url?.includes(url));

    if (!shouldAttemptRefresh) {
      return Promise.reject(error);
    }

    originalRequest._retry = true;

    try {
      refreshPromise ??= refreshAccessToken().finally(() => {
        refreshPromise = null;
      });
      const token = await refreshPromise;
      useAuthStore.getState().setToken(token);
      originalRequest.headers.Authorization = `Bearer ${token}`;
      return api(originalRequest);
    } catch {
      // Refresh token is missing/expired — the session is truly over.
      useAuthStore.getState().setToken(null);
      return Promise.reject(error);
    }
  }
);

interface ClientConfig {
  /** Distance band thresholds in km, ascending. Configured server-side. */
  distanceBands: number[];
}

// Fetched rather than hardcoded so the distance filter always offers the same
// steps the server actually bands by.
export const getClientConfig = async (): Promise<ClientConfig> => {
  const response = await api.get<ClientConfig>("/api/v1/config");
  return response.data;
};

// Mirror of the backend upload rules (see profile handler): keep in sync.
export const PHOTO_MAX_SIZE_MB = 10;
export const PHOTO_ALLOWED_TYPES = ["image/jpeg", "image/png", "image/webp"];
export const PHOTO_ACCEPT = PHOTO_ALLOWED_TYPES.join(",");
export const PHOTO_RULES_HINT = `JPEG, PNG or WebP, up to ${PHOTO_MAX_SIZE_MB}MB`;

export const uploadPhoto = async (file: File) => {
  // Pre-validate so the user gets instant feedback instead of a server error.
  // The backend still enforces both rules by sniffing the actual bytes.
  if (!PHOTO_ALLOWED_TYPES.includes(file.type)) {
    throw new Error(`Unsupported file type. ${PHOTO_RULES_HINT}.`);
  }
  if (file.size > PHOTO_MAX_SIZE_MB * 1024 * 1024) {
    throw new Error(`File is too large. ${PHOTO_RULES_HINT}.`);
  }

  const formData = new FormData();
  formData.append("photo", file);

  const response = await api.post<{ key: string; url: string }>(
    "/api/v1/profiles/upload",
    formData,
    {
      headers: {
        "Content-Type": "multipart/form-data",
      },
    }
  );
  return response.data;
};
