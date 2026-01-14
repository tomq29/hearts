import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useNavigate } from '@tanstack/react-router'
import { useMutation } from '@tanstack/react-query'
import { api } from '@/shared/api'
import { useAuthStore } from '@/entities/session'
import { getErrorMessage } from '@/shared/lib/error'
import { loginSchema, type LoginFormSchema } from '../model/schema'

export const LoginForm = () => {
  const navigate = useNavigate()
  const setToken = useAuthStore((state) => state.setToken)

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormSchema>({
    resolver: zodResolver(loginSchema),
  })

  const mutation = useMutation({
    mutationFn: (data: LoginFormSchema) => api.post('/api/v1/users/login', data),
    onSuccess: (response) => {
      const token = response.data.accessToken
      setToken(token)
      navigate({ to: '/' })
    },
  })

  const onSubmit = (data: LoginFormSchema) => {
    mutation.mutate(data)
  }

  return (
    <div className="w-full max-w-md p-8 bg-white rounded-xl shadow-lg border border-gray-100">
      <h2 className="text-2xl font-bold text-center mb-6 text-gray-800">
        Welcome Back
      </h2>

      {mutation.isError && (
        <div className="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-lg border border-red-100">
          {getErrorMessage(mutation.error)}
        </div>
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Email
          </label>
          <input
            {...register('email')}
            type="email"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition"
            placeholder="you@example.com"
          />
          {errors.email && (
            <p className="text-red-500 text-xs mt-1">
              {errors.email.message}
            </p>
          )}
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">
            Password
          </label>
          <input
            {...register('password')}
            type="password"
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-pink-500 focus:border-transparent outline-none transition"
            placeholder="••••••••"
          />
          {errors.password && (
            <p className="text-red-500 text-xs mt-1">
              {errors.password.message}
            </p>
          )}
        </div>

        <button
          type="submit"
          disabled={mutation.isPending}
          className="w-full bg-pink-600 text-white py-2.5 rounded-lg font-medium hover:bg-pink-700 transition disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {mutation.isPending ? 'Signing in...' : 'Sign In'}
        </button>
      </form>
      
      <div className="mt-6 text-center text-sm text-gray-500">
        Don't have an account?{' '}
        <a href="/register" className="text-pink-600 hover:underline font-medium">
          Sign up
        </a>
      </div>
    </div>
  )
}
