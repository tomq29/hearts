import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { useAuthStore } from "@/entities/session";

const BASE_URL = "http://localhost:8080"; // Adjust if your API is on a different port

export const api = axios.create({
  baseURL: BASE_URL,
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
    `${BASE_URL}/api/v1/auth/refresh`,
    null,
    { withCredentials: true }
  );
  return response.data.accessToken;
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

export const uploadPhoto = async (file: File) => {
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
