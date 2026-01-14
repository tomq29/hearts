import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/matches')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/matches"!</div>
}
