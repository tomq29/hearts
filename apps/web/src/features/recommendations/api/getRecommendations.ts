import { api } from "@/shared/api";
import type { Profile } from "@/entities/profile";

export const getRecommendations = async ({
  pageParam = 1,
}: {
  pageParam?: number;
}) => {
  const limit = 10;
  const response = await api.get<Profile[]>("/api/v1/profiles", {
    params: {
      page: pageParam,
      limit,
    },
  });
  return response.data;
};
