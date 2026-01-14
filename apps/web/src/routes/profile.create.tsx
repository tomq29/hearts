import { createFileRoute } from '@tanstack/react-router'
import { ProfileCreatePage } from '@/pages/profile-create/ui/ProfileCreatePage'

export const Route = createFileRoute('/profile/create')({
  component: ProfileCreatePage,
})
