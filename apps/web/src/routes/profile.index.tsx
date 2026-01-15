import { createFileRoute } from '@tanstack/react-router'
import { ProfilePage } from '@/pages/profile/ui/ProfilePage'

export const Route = createFileRoute('/profile/')({
  component: ProfilePage,
})
