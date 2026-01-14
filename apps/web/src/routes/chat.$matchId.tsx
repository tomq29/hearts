import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/chat/$matchId')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/chat/$matchId"!</div>
}
