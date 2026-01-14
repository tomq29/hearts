import type { Profile } from '../model/types'

interface ProfileCardProps {
  profile: Profile;
}

export const ProfileCard = ({ profile }: ProfileCardProps) => {
  const age =
    new Date().getFullYear() - new Date(profile.birthDate).getFullYear();

  return (
    <div className="bg-white rounded-xl shadow-sm overflow-hidden border border-gray-100 max-w-sm mx-auto hover:shadow-md transition-shadow">
      <div className="aspect-[3/4] relative bg-gray-100">
        {profile.photos[0] ? (
          <img
            src={profile.photos[0]}
            alt={profile.firstName}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-gray-400">
            No Photo
          </div>
        )}
        <div className="absolute bottom-0 left-0 right-0 bg-gradient-to-t from-black/60 to-transparent p-4 pt-12 text-white">
          <h3 className="text-2xl font-bold">
            {profile.firstName}, {age}
          </h3>
          {profile.height && (
            <p className="text-sm opacity-90">{profile.height} cm</p>
          )}
        </div>
      </div>

      <div className="p-4">
        {profile.bio && (
          <p className="text-gray-600 text-sm line-clamp-3">{profile.bio}</p>
        )}

        <div className="mt-4 flex gap-2">
          <span className="px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded-full capitalize">
            {profile.gender}
          </span>
        </div>
      </div>
    </div>
  );
};
