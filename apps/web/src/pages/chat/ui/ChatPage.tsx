import { useParams } from "@tanstack/react-router";
import { ChatWindow } from "@/features/chat";

export const ChatPage = () => {
  const { matchId } = useParams({ from: "/chat/$matchId" });

  return (
    <div className="max-w-4xl mx-auto p-4 h-[calc(100vh-4rem)] flex flex-col">
      <div className="mb-4">
        <h1 className="text-2xl font-bold text-gray-800">Chat</h1>
      </div>
      <div className="flex-1 min-h-0 border rounded-lg shadow-sm bg-white overflow-hidden">
        <ChatWindow matchId={matchId} />
      </div>
    </div>
  );
};
