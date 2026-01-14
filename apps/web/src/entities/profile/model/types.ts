export interface Profile {
  id: string
  userId: string
  firstName: string
  birthDate: string
  gender: 'male' | 'female' | 'other'
  height?: number
  bio?: string
  photos: string[]
}
