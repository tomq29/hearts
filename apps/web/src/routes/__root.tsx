import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { useAuthStore } from "@/entities/session";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  const token = useAuthStore((state) => state.token);
  const isAuthenticated = !!token;

  return (
    <>
      <div className="p-2 flex gap-2 text-lg border-b border-gray-200">
        <Link
          to="/"
          activeProps={{
            className: "font-bold",
          }}
          activeOptions={{ exact: true }}
        >
          Home
        </Link>
        {!isAuthenticated && (
          <>
            <Link
              to="/login"
              activeProps={{
                className: "font-bold",
              }}
            >
              Login
            </Link>
            <Link
              to="/register"
              activeProps={{
                className: "font-bold",
              }}
            >
              Register
            </Link>
          </>
        )}
        {isAuthenticated && (
          <>
            <Link
              to="/matches"
              activeProps={{
                className: "font-bold",
              }}
            >
              Matches
            </Link>
            <Link
              to="/profile"
              activeProps={{
                className: "font-bold",
              }}
            >
              Profile
            </Link>
          </>
        )}
      </div>
      <hr />
      <div className="p-2">
        <Outlet />
      </div>
      <TanStackRouterDevtools position="bottom-right" />
    </>
  );
}
