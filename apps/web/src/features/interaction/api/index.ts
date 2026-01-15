import { api } from "@/shared/api";

export const sendInteraction = async (
  targetId: string,
  type: "like" | "pass"
) => {
  // Backend expects payload: { targetId: string, isLike: boolean }
  const isLike = type === "like";
  const response = await api.post("/api/v1/likes", {
    targetId,
    isLike,
  });
  return response.data;
};
