// packages/types/src/index.ts

export interface User {
  id: string;
  email: string;
  username: string;
  createdAt: Date;
}

export interface Profile {
  id: string;
  userId: string; // Links to the User
  firstName: string;
  bio: string;
  photos: string[];

  // The core feature of your app
  selfDescribedFlaws: string[];

  // It might be good to balance the "flaws" with strengths
  selfDescribedStrengths: string[];

  createdAt: Date;
  updatedAt: Date;
}
