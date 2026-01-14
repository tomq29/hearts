import { api } from "@/shared/api";
import type { Profile } from "../model/types";

export const getMyProfile = async () => {
  const response = await api.get<Profile>("/api/v1/profiles/me");
  return response.data;
};

export const updateProfile = async (data: Partial<Profile>) => {
  const response = await api.put<Profile>("/api/v1/profiles/me", data);
  return response.data;
};
