import { api } from "@/shared/api";
import type { Message } from "../model/types";

export const getMessages = async (matchId: string) => {
  const response = await api.get<Message[]>(
    `/api/v1/matches/${matchId}/messages`
  );
  return response.data;
};

export const sendMessage = async (matchId: string, content: string) => {
  const response = await api.post<Message>(
    `/api/v1/matches/${matchId}/messages`,
    {
      content,
    }
  );
  return response.data;
};

export const getChatTicket = async () => {
  const response = await api.post<{ ticket: string }>("/api/v1/chat/ticket");
  return response.data.ticket;
};
