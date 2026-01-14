import { api } from "@/shared/api";
import type { Profile } from "@/entities/profile";

export interface Match {
  id: string;
  profile: Profile;
  lastMessage?: string;
  createdAt: string;
}

export const getMatches = async () => {
  const response = await api.get<Match[]>("/api/v1/matches");
  return response.data;
};
