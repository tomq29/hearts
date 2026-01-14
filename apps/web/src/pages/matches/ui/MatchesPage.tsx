import { MatchesList } from "@/features/matches";

export const MatchesPage = () => {
  return (
    <div className="max-w-4xl mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6 text-gray-800">Your Matches</h1>
      <MatchesList />
    </div>
  );
};
