import { useQuery } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { getMatches } from "../api";
import { getErrorMessage } from "@/shared/lib/error";

export const MatchesList = () => {
  const {
    data: matches,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["matches"],
    queryFn: getMatches,
  });

  if (isLoading) {
    return (
      <div className="flex justify-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-pink-500"></div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="text-center p-8 text-red-500">
        <p>Failed to load matches</p>
        <p className="text-sm mt-2 text-gray-500">{getErrorMessage(error)}</p>
      </div>
    );
  }

  if (!matches || matches.length === 0) {
    return (
      <div className="text-center p-12 bg-gray-50 rounded-xl border border-dashed border-gray-200">
        <p className="text-gray-500 text-lg">No matches yet 💔</p>
        <p className="text-gray-400 text-sm mt-2">
          Keep swiping to find your perfect match!
        </p>
      </div>
    );
  }

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
      {matches.map((match) => (
        <Link
          key={match.id}
          to="/chat/$matchId"
          params={{ matchId: match.id }}
          className="flex items-center gap-4 p-4 bg-white rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition hover:border-pink-200 group"
        >
          <div className="relative w-16 h-16 rounded-full overflow-hidden bg-gray-100 flex-shrink-0">
            {match.profile.photos[0] ? (
              <img
                src={match.profile.photos[0]}
                alt={match.profile.firstName}
                className="w-full h-full object-cover group-hover:scale-110 transition duration-500"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center text-gray-400 text-xs">
                No Photo
              </div>
            )}
          </div>

          <div className="flex-1 min-w-0">
            <h3 className="font-semibold text-gray-900 truncate">
              {match.profile.firstName}
            </h3>
            <p className="text-sm text-gray-500 truncate">
              {match.lastMessage || "Start a conversation 👋"}
            </p>
          </div>

          <div className="w-2 h-2 rounded-full bg-pink-500 opacity-0 group-hover:opacity-100 transition"></div>
        </Link>
      ))}
    </div>
  );
};
