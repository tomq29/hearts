import { createFileRoute } from "@tanstack/react-router";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/shared/api";
import type { Profile } from "@/entities/profile/model/types";

export const Route = createFileRoute("/matches")({
  component: MatchesPage,
});

function MatchesPage() {
  const {
    data: matches,
    isLoading,
    error,
  } = useQuery({
    queryKey: ["matches"],
    queryFn: async () => {
      const response = await api.get<Profile[]>("/api/v1/matches");
      return response.data;
    },
  });

  if (isLoading) {
    return (
      <div className="p-8 text-center text-gray-500">Loading matches...</div>
    );
  }

  if (error) {
    return (
      <div className="p-8 text-center text-red-500">Failed to load matches</div>
    );
  }

  if (!matches || matches.length === 0) {
    return (
      <div className="p-8 text-center">
        <h2 className="text-2xl font-bold mb-4">No matches yet</h2>
        <p className="text-gray-600">Keep swiping to find your match!</p>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 max-w-4xl">
      <h1 className="text-3xl font-bold mb-8">Your Matches</h1>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
        {matches.map((profile) => (
          <div
            key={profile.id}
            className="relative aspect-[3/4] rounded-xl overflow-hidden shadow-md group"
          >
            {profile.photos?.[0] ? (
              <img
                src={profile.photos[0]}
                alt={profile.firstName}
                className="absolute inset-0 w-full h-full object-cover"
              />
            ) : (
              <div className="absolute inset-0 bg-gray-200 flex items-center justify-center text-gray-400">
                No Photo
              </div>
            )}
            <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent flex flex-col justify-end p-4 text-white">
              <h3 className="font-bold text-lg">{profile.firstName}</h3>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
