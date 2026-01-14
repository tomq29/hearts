import { useEffect, useRef, useCallback, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useAuthStore } from "@/entities/session";
import type { Message } from "../model/types";
import { getChatTicket } from "../api";

export const useChatSocket = (matchId: string) => {
  const socketRef = useRef<WebSocket | null>(null);
  const queryClient = useQueryClient();
  const token = useAuthStore((state) => state.token);
  const [isPartnerTyping, setIsPartnerTyping] = useState(false);

  useEffect(() => {
    if (!token || !matchId) return;

    let isMounted = true;
    let reconnectTimeout: ReturnType<typeof setTimeout>;
    let typingTimeout: ReturnType<typeof setTimeout>;

    const connect = async () => {
      try {
        const ticket = await getChatTicket();

        if (!isMounted) return;

        const wsUrl = `ws://localhost:8080/ws?ticket=${ticket}&matchId=${matchId}`;
        const socket = new WebSocket(wsUrl);

        socket.onopen = () => {
          console.log("Connected to chat");
        };

        socket.onmessage = (event) => {
          try {
            const payload = JSON.parse(event.data);

            if (payload.type === "typing") {
              setIsPartnerTyping(true);
              clearTimeout(typingTimeout);
              typingTimeout = setTimeout(() => setIsPartnerTyping(false), 3000);
              return;
            }

            // Assume it's a message if not typing
            const message: Message = payload;

            queryClient.setQueryData<Message[]>(
              ["messages", matchId],
              (old) => {
                if (!old) return [message];
                if (old.find((m) => m.id === message.id)) return old;
                return [...old, message];
              }
            );
          } catch (error) {
            console.error("Failed to parse message", error);
          }
        };

        socket.onclose = () => {
          console.log("Disconnected from chat");
          if (isMounted) {
            console.log("Attempting to reconnect in 3s...");
            reconnectTimeout = setTimeout(connect, 3000);
          }
        };

        socket.onerror = (error) => {
          console.error("Socket error", error);
          socket.close();
        };

        socketRef.current = socket;
      } catch (error) {
        console.error("Failed to connect to chat", error);
        if (isMounted) {
          reconnectTimeout = setTimeout(connect, 3000);
        }
      }
    };

    connect();

    return () => {
      isMounted = false;
      clearTimeout(reconnectTimeout);
      clearTimeout(typingTimeout);
      if (socketRef.current) {
        socketRef.current.onclose = null;
        socketRef.current.close();
      }
    };
  }, [matchId, token, queryClient]);

  const sendMessageSocket = useCallback((content: string) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ content }));
    } else {
      console.warn("Socket not connected");
    }
  }, []);

  const sendTyping = useCallback((toUserId: string) => {
    if (socketRef.current && socketRef.current.readyState === WebSocket.OPEN) {
      socketRef.current.send(JSON.stringify({ type: "typing", toUserId }));
    }
  }, []);

  return { sendMessageSocket, sendTyping, isPartnerTyping };
};
