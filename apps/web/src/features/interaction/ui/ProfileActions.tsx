import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { sendInteraction } from "../api";
import { CrossIcon, HeartIcon } from "@/shared/ui/icons";

interface ProfileActionsProps {
  targetUserId: string;
  onAction?: () => void;
  initialInteractionType?: "like" | "pass" | null;
}

export const ProfileActions = ({
  targetUserId,
  onAction,
  initialInteractionType,
}: ProfileActionsProps) => {
  const queryClient = useQueryClient();
  const [interactionState, setInteractionState] = useState<
    "like" | "pass" | null
  >(initialInteractionType || null);

  const mutation = useMutation({
    mutationFn: ({ type }: { type: "like" | "pass" }) =>
      sendInteraction(targetUserId, type),
    onSuccess: (_, variables) => {
      setInteractionState(variables.type);
      queryClient.invalidateQueries({ queryKey: ["recommendations"] });
      onAction?.();
    },
  });

  const isLike = interactionState === "like";
  const isPass = interactionState === "pass";
  const hasInteracted = isLike || isPass;

  return (
    <div className="flex justify-center gap-4 mt-4">
      <button
        onClick={() => mutation.mutate({ type: "pass" })}
        disabled={mutation.isPending || hasInteracted}
        className={`w-12 h-12 flex items-center justify-center rounded-full border transition shadow-sm 
          ${
            isPass
              ? "bg-red-100 border-red-200 text-red-500 cursor-default"
              : hasInteracted
                ? "bg-white border-gray-100 text-gray-200 cursor-not-allowed opacity-50"
                : "bg-white border-gray-200 text-gray-400 hover:bg-gray-50 hover:text-gray-600 hover:border-gray-300"
          }`}
        aria-label="Pass"
      >
        <CrossIcon />
      </button>

      <button
        onClick={() => mutation.mutate({ type: "like" })}
        disabled={mutation.isPending || hasInteracted}
        className={`w-12 h-12 flex items-center justify-center rounded-full transition shadow-md 
          ${
            isLike
              ? "bg-pink-600 text-white cursor-default"
              : hasInteracted
                ? "bg-gray-200 text-gray-400 cursor-not-allowed opacity-50"
                : "bg-pink-500 text-white hover:bg-pink-600 hover:scale-105"
          }`}
        aria-label="Like"
      >
        <HeartIcon />
      </button>
    </div>
  );
};
