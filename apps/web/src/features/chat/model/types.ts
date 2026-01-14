export interface Message {
  id: string;
  matchId: string;
  senderId: string;
  content: string;
  createdAt: string;
}

export interface SendMessagePayload {
  matchId: string;
  content: string;
}
