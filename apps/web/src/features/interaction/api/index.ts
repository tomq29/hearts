import { api } from "@/shared/api";

export const sendInteraction = async (
  targetId: string,
  type: "like" | "pass"
) => {
  const response = await api.post("/api/v1/interactions", {
    targetId,
    type,
  });
  return response.data;
};
