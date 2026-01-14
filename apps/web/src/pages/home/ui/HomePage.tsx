import { RecommendationFeed } from "@/features/recommendations";

export function HomePage() {
  return (
    <div className="max-w-6xl mx-auto">
      <div className="p-6 mb-6 bg-gradient-to-r from-pink-500 to-purple-600 rounded-b-3xl shadow-lg text-white">
        <h1 className="text-3xl font-bold mb-2">Discover</h1>
        <p className="opacity-90">Find your perfect match nearby</p>
      </div>

      <RecommendationFeed />
    </div>
  );
}
