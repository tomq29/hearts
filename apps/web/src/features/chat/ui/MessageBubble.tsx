import { cn } from "@/shared/lib/utils";
import type { Message } from "../model/types";

interface MessageBubbleProps {
  message: Message;
  isMine: boolean;
}

export const MessageBubble = ({ message, isMine }: MessageBubbleProps) => {
  return (
    <div
      className={cn("flex w-full", isMine ? "justify-end" : "justify-start")}
    >
      <div
        className={cn(
          "max-w-[70%] px-4 py-2 rounded-2xl text-sm",
          isMine
            ? "bg-pink-500 text-white rounded-br-none"
            : "bg-gray-100 text-gray-800 rounded-bl-none"
        )}
      >
        <p>{message.content}</p>
        <span
          className={cn(
            "text-[10px] block mt-1 opacity-70",
            isMine ? "text-pink-100" : "text-gray-400"
          )}
        >
          {new Date(message.createdAt).toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </span>
      </div>
    </div>
  );
};
