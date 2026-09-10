export interface Profile {
  id: string
  userId: string
  firstName: string
  birthDate: string
  gender: 'male' | 'female' | 'other'
  height?: number
  bio?: string
  photos: string[]
  interactionType?: 'like' | 'pass' | null
  /** Coarse distance from you, e.g. "5–10 km". Absent when either of you has no location set. */
  distanceBand?: string
}
