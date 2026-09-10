import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import axios from "axios";
import { getRecommendations } from "../api/getRecommendations";
import { ProfileCard } from "@/entities/profile";
import { getClientConfig } from "@/shared/api";
import { getErrorMessage } from "@/shared/lib/error";
import { useIntersection } from "@/shared/lib/hooks/useIntersection";
import { ProfileActions } from "@/features/interaction";

const RADIUS_STORAGE_KEY = "feed-radius";

// Remembering the choice is a per-viewer convenience, so a blocked or
// unavailable localStorage just means the filter starts at "Any".
const readStoredRadius = (): number | undefined => {
  try {
    const stored = localStorage.getItem(RADIUS_STORAGE_KEY);
    return stored ? Number(stored) : undefined;
  } catch {
    return undefined;
  }
};

export const RecommendationFeed = () => {
  const [radius, setRadius] = useState<number | undefined>(readStoredRadius);

  // Filter steps come from the server so they match the bands shown on cards.
  const { data: config } = useQuery({
    queryKey: ["client-config"],
    queryFn: getClientConfig,
    staleTime: Infinity,
  });

  const handleRadiusChange = (value: number | undefined) => {
    setRadius(value);
    try {
      if (value) {
        localStorage.setItem(RADIUS_STORAGE_KEY, String(value));
      } else {
        localStorage.removeItem(RADIUS_STORAGE_KEY);
      }
    } catch {
      // Not persisting the choice is not worth interrupting the user for.
    }
  };

  const {
    data,
    isLoading,
    isError,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    queryKey: ["recommendations", radius],
    queryFn: ({ pageParam }) => getRecommendations({ pageParam, radius }),
    initialPageParam: 1,
    getNextPageParam: (lastPage, allPages) => {
      // Assuming the API returns an empty array when there are no more items
      // Or we could check if lastPage.length < limit
      return lastPage.length === 10 ? allPages.length + 1 : undefined;
    },
  });

  const { ref, isIntersecting } = useIntersection({
    rootMargin: "100px",
  });

  useEffect(() => {
    if (isIntersecting && hasNextPage) {
      fetchNextPage();
    }
  }, [isIntersecting, hasNextPage, fetchNextPage]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-pink-500"></div>
      </div>
    );
  }

  if (isError) {
    // Check if error is 404 (Profile Not Found)
    if (axios.isAxiosError(error) && error.response?.status === 404) {
      return (
        <div className="text-center p-8">
          <h3 className="text-xl font-semibold mb-3">Profile Required</h3>
          <p className="text-gray-600 mb-6">
            You need to create a profile before you can see other people.
          </p>
          <Link
            to="/profile/create"
            className="bg-pink-500 text-white px-6 py-3 rounded-full font-medium hover:bg-pink-600 transition-colors"
          >
            Create Profile
          </Link>
        </div>
      );
    }

    return (
      <div className="text-center p-8 text-red-500">
        <p>Failed to load recommendations</p>
        <p className="text-sm mt-2 text-gray-500">{getErrorMessage(error)}</p>
      </div>
    );
  }

  const profiles = data?.pages.flat() || [];

  // Rendered in both the empty and populated states — a filter you cannot
  // widen after it returns nothing is a dead end.
  const distanceFilter = config?.distanceBands.length ? (
    <div className="flex items-center gap-2 px-4">
      <label htmlFor="radius" className="text-sm text-gray-600">
        Distance
      </label>
      <select
        id="radius"
        value={radius ?? ""}
        onChange={(e) =>
          handleRadiusChange(e.target.value ? Number(e.target.value) : undefined)
        }
        className="px-3 py-1.5 border border-gray-300 rounded-lg text-sm outline-none focus:ring-2 focus:ring-pink-500"
      >
        <option value="">Any distance</option>
        {config.distanceBands.map((km) => (
          <option key={km} value={km}>
            Within {km} km
          </option>
        ))}
      </select>
    </div>
  ) : null;

  if (profiles.length === 0) {
    return (
      <div className="space-y-6 pt-4">
        {distanceFilter}
        <div className="text-center p-8 text-gray-500">
          <p>
            {radius
              ? `No profiles found within ${radius} km. Try a wider distance.`
              : "No profiles found nearby."}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6 pb-8 pt-4">
      {distanceFilter}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6 p-4">
        {profiles.map((profile, index) => (
          <div key={`${profile.id}-${index}`} className="flex flex-col">
            <ProfileCard profile={profile} />
            <ProfileActions 
              targetUserId={profile.userId} 
              initialInteractionType={profile.interactionType}
            />
          </div>
        ))}
      </div>

      {/* Intersection target for infinite scroll */}
      <div ref={ref} className="h-10 flex justify-center items-center">
        {isFetchingNextPage && (
          <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-pink-500"></div>
        )}
      </div>
    </div>
  );
};
