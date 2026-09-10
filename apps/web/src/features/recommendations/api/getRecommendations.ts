import { api } from "@/shared/api";
import type { Profile } from "@/entities/profile";

export const getRecommendations = async ({
  pageParam = 1,
  radius,
}: {
  pageParam?: number;
  radius?: number;
}) => {
  const limit = 10;
  const response = await api.get<Profile[]>("/api/v1/profiles/search", {
    params: {
      page: pageParam,
      limit,
      // Omitted for "Any distance". The server also ignores it when the
      // viewer has no location set, rather than failing.
      ...(radius ? { radius } : {}),
    },
  });
  return response.data;
};
