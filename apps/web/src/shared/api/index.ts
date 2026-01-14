import axios from "axios";
import { useAuthStore } from "@/entities/session";

export const api = axios.create({
  baseURL: "http://localhost:8080", // Adjust if your API is on a different port
});

api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

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
