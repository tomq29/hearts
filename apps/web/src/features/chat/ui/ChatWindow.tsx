import { useState, useEffect, useRef } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { getMessages, sendMessage } from "../api";
import { useChatSocket } from "../lib/useChatSocket";
import { MessageBubble } from "./MessageBubble";
import { getErrorMessage } from "@/shared/lib/error";
import { getMyProfile } from "@/entities/profile";
import type { Message } from "../model/types";

interface ChatWindowProps {
  matchId: string;
}

export const ChatWindow = ({ matchId }: ChatWindowProps) => {
  const [inputValue, setInputValue] = useState("");
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const queryClient = useQueryClient();

  // Fetch my profile to know my ID (to determine isMine)
  const { data: myProfile } = useQuery({
    queryKey: ["my-profile"],
    queryFn: getMyProfile,
  });

  const {
    data: messages,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["messages", matchId],
    queryFn: () => getMessages(matchId),
    refetchInterval: false, // We rely on WebSocket for updates
  });

  // Initialize WebSocket
  const { sendTyping, isPartnerTyping } = useChatSocket(matchId);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isPartnerTyping]);

  const mutation = useMutation({
    mutationFn: (content: string) => sendMessage(matchId, content),
    onSuccess: (newMessage) => {
      setInputValue("");
      // Optimistically update or wait for socket?
      // Usually REST response is faster/reliable for the sender.
      queryClient.setQueryData<Message[]>(["messages", matchId], (old) => {
        if (!old) return [newMessage];
        return [...old, newMessage];
      });
    },
  });

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim()) return;
    mutation.mutate(inputValue);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
    // We need the other user's ID to send the typing event.
    // In a real app, we'd get this from the match details or the messages.
    // For now, let's assume we can get it from the first message that isn't ours,
    // or we need to fetch match details.
    // Let's assume we have a way to get the partner ID.
    // For this MVP, we might need to fetch match details to get the partner ID.
    // But wait, the backend needs the target UserID.
    // Let's try to find the partner ID from the messages list if available.
    const partnerId = messages?.find(
      (m) => m.senderId !== myProfile?.userId
    )?.senderId;
    if (partnerId) {
      sendTyping(partnerId);
    }
  };

  if (isLoading) {
    return <div className="flex justify-center p-8">Loading chat...</div>;
  }

  if (isError) {
    return (
      <div className="text-center p-8 text-red-500">
        {getErrorMessage(error)}
      </div>
    );
  }

  return (
    <div className="flex flex-col h-[calc(100vh-200px)] bg-white rounded-xl shadow-sm border border-gray-100 overflow-hidden">
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages?.map((msg) => (
          <MessageBubble
            key={msg.id}
            message={msg}
            isMine={msg.senderId === myProfile?.userId}
          />
        ))}
        {isPartnerTyping && (
          <div className="flex w-full justify-start">
            <div className="bg-gray-100 text-gray-500 px-4 py-2 rounded-2xl rounded-bl-none text-sm italic animate-pulse">
              Typing...
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form
        onSubmit={handleSend}
        className="p-4 border-t border-gray-100 bg-gray-50 flex gap-2"
      >
        <input
          type="text"
          value={inputValue}
          onChange={handleInputChange}
          placeholder="Type a message..."
          className="flex-1 px-4 py-2 rounded-full border border-gray-300 focus:outline-none focus:border-pink-500 focus:ring-1 focus:ring-pink-500"
        />
        <button
          type="submit"
          disabled={mutation.isPending || !inputValue.trim()}
          className="bg-pink-500 text-white px-6 py-2 rounded-full font-medium hover:bg-pink-600 transition disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Send
        </button>
      </form>
    </div>
  );
};
