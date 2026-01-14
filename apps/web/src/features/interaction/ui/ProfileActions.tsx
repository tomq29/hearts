import { useMutation, useQueryClient } from '@tanstack/react-query'
import { sendInteraction } from '../api'
import { CrossIcon, HeartIcon } from '@/shared/ui/icons'

interface ProfileActionsProps {
  profileId: string
  onAction?: () => void
}

export const ProfileActions = ({ profileId, onAction }: ProfileActionsProps) => {
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: ({ type }: { type: 'like' | 'pass' }) => sendInteraction(profileId, type),
    onSuccess: () => {
      // Invalidate recommendations to refresh the list if needed, 
      // but ideally we just remove the item locally via onAction
      queryClient.invalidateQueries({ queryKey: ['recommendations'] })
      onAction?.()
    },
  })

  return (
    <div className="flex justify-center gap-4 mt-4">
      <button
        onClick={() => mutation.mutate({ type: 'pass' })}
        disabled={mutation.isPending}
        className="w-12 h-12 flex items-center justify-center rounded-full bg-white border border-gray-200 text-gray-400 hover:bg-gray-50 hover:text-gray-600 hover:border-gray-300 transition shadow-sm disabled:opacity-50"
        aria-label="Pass"
      >
        <CrossIcon />
      </button>

      <button
        onClick={() => mutation.mutate({ type: 'like' })}
        disabled={mutation.isPending}
        className="w-12 h-12 flex items-center justify-center rounded-full bg-pink-500 text-white hover:bg-pink-600 hover:scale-105 transition shadow-md disabled:opacity-50"
        aria-label="Like"
      >
        <HeartIcon />
      </button>
    </div>
  )
}
