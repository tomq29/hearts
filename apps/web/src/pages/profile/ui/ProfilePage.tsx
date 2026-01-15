import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { getMyProfile, ProfileCard } from "@/entities/profile";
import { EditProfileForm } from "@/features/profile/edit-form";
import { CreateProfileForm } from "@/features/profile/create-form/ui/CreateProfileForm";
import { getErrorMessage } from "@/shared/lib/error";
import { useAuthStore } from "@/entities/session";

export const ProfilePage = () => {
  const navigate = useNavigate();
  const logout = useAuthStore((state) => state.setToken);
  const [isEditing, setIsEditing] = useState(false);

  const {
    data: profile,
    isLoading,
    isError,
    error,
  } = useQuery({
    queryKey: ["my-profile"],
    queryFn: getMyProfile,
    retry: false, // Don't retry if 404
  });

  const handleLogout = () => {
    logout(null);
    navigate({ to: "/login" });
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center min-h-[50vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-pink-500"></div>
      </div>
    );
  }

  // Handle 404 specifically - Show Create Profile Form
  // We assume that if it's a 404, it means the user needs to create a profile.
  if (isError) {
    const errorMsg = getErrorMessage(error);
    if (errorMsg.includes("404") || errorMsg.includes("not found")) {
      return <CreateProfileForm />;
    }

    return (
      <div className="text-center p-8 text-red-500">
        <p>Failed to load profile</p>
        <p className="text-sm mt-2 text-gray-500">{errorMsg}</p>
      </div>
    );
  }

  if (!profile) {
    return <CreateProfileForm />;
  }

  return (
    <div className="max-w-4xl mx-auto p-4">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-3xl font-bold text-gray-800">My Profile</h1>
        <div className="flex gap-2">
          <button
            onClick={() => setIsEditing(!isEditing)}
            className="px-4 py-2 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 transition"
          >
            {isEditing ? "Cancel" : "Edit Profile"}
          </button>
          <button
            onClick={handleLogout}
            className="px-4 py-2 bg-red-50 text-red-600 rounded-lg hover:bg-red-100 transition"
          >
            Logout
          </button>
        </div>
      </div>

      {isEditing ? (
        <EditProfileForm />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          <div className="md:col-span-1">
            <ProfileCard profile={profile} />
          </div>
          <div className="md:col-span-2 space-y-6">
            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
              <h2 className="text-xl font-semibold mb-4">About Me</h2>
              <p className="text-gray-600 whitespace-pre-wrap">
                {profile.bio || "No bio provided."}
              </p>
            </div>

            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
              <h2 className="text-xl font-semibold mb-4">Details</h2>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <span className="block text-sm text-gray-500">Gender</span>
                  <span className="capitalize">{profile.gender}</span>
                </div>
                <div>
                  <span className="block text-sm text-gray-500">Height</span>
                  <span>{profile.height ? `${profile.height} cm` : "-"}</span>
                </div>
                <div>
                  <span className="block text-sm text-gray-500">Age</span>
                  <span>
                    {new Date().getFullYear() -
                      new Date(profile.birthDate).getFullYear()}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
