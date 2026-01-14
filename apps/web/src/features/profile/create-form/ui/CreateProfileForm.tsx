import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { api, uploadPhoto } from '@/shared/api'
import { getErrorMessage } from '@/shared/lib/error'
import { profileSchema, type ProfileFormSchema } from '@/entities/profile'

export const CreateProfileForm = () => {
  const navigate = useNavigate()
  const [uploading, setUploading] = useState(false)
  const [photoUrls, setPhotoUrls] = useState<string[]>([])

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<ProfileFormSchema>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      photos: [],
    },
  })

  const photos = watch('photos')

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return

    try {
      setUploading(true)
      const file = e.target.files[0]
      const { key, url } = await uploadPhoto(file)
      
      setValue('photos', [...photos, key])
      setPhotoUrls([...photoUrls, url])
    } catch (err) {
      console.error('Upload failed', err)
      alert('Failed to upload photo')
    } finally {
      setUploading(false)
    }
  }

  const mutation = useMutation({
    mutationFn: (data: ProfileFormSchema) => api.post('/api/v1/profiles', {
      ...data,
      birthDate: new Date(data.birthDate).toISOString(),
      height: data.height ? Number(data.height) : undefined,
    }),
    onSuccess: () => {
      navigate({ to: '/' })
    },
  })

  const onSubmit = (data: ProfileFormSchema) => {
    mutation.mutate(data)
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-8 text-gray-800">Create Your Profile</h1>
      
      {mutation.isError && (
        <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-100">
          {getErrorMessage(mutation.error)}
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        {/* Photos Section */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">Photos</label>
          <div className="flex flex-wrap gap-4">
            {photoUrls.map((url, index) => (
              <div key={index} className="relative w-24 h-32 rounded-lg overflow-hidden border border-gray-200">
                <img src={url} alt="Profile" className="w-full h-full object-cover" />
              </div>
            ))}
            
            <label className="w-24 h-32 flex flex-col items-center justify-center border-2 border-dashed border-gray-300 rounded-lg cursor-pointer hover:border-pink-500 hover:bg-pink-50 transition">
              <span className="text-2xl text-gray-400">+</span>
              <span className="text-xs text-gray-500 mt-1">{uploading ? '...' : 'Add'}</span>
              <input type="file" className="hidden" accept="image/*" onChange={handleFileChange} disabled={uploading} />
            </label>
          </div>
          {errors.photos && <p className="text-red-500 text-xs mt-1">{errors.photos.message}</p>}
        </div>

        {/* Basic Info */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">First Name</label>
            <input
              {...register('firstName')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
              placeholder="Your name"
            />
            {errors.firstName && <p className="text-red-500 text-xs mt-1">{errors.firstName.message}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Birth Date</label>
            <input
              type="date"
              {...register('birthDate')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
            />
            {errors.birthDate && <p className="text-red-500 text-xs mt-1">{errors.birthDate.message}</p>}
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Gender</label>
            <select
              {...register('gender')}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none bg-white"
            >
              <option value="">Select gender</option>
              <option value="male">Male</option>
              <option value="female">Female</option>
              <option value="other">Other</option>
            </select>
            {errors.gender && <p className="text-red-500 text-xs mt-1">{errors.gender.message}</p>}
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Height (cm)</label>
            <input
              type="number"
              {...register('height', { valueAsNumber: true })}
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none"
              placeholder="175"
            />
            {errors.height && <p className="text-red-500 text-xs mt-1">{errors.height.message}</p>}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Bio</label>
          <textarea
            {...register('bio')}
            rows={4}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 outline-none resize-none"
            placeholder="Tell us about yourself..."
          />
          {errors.bio && <p className="text-red-500 text-xs mt-1">{errors.bio.message}</p>}
        </div>

        <button
          type="submit"
          disabled={mutation.isPending}
          className="w-full bg-pink-600 text-white py-3 rounded-lg font-medium hover:bg-pink-700 transition disabled:opacity-50"
        >
          {mutation.isPending ? 'Creating Profile...' : 'Complete Profile'}
        </button>
      </form>
    </div>
  )
}
