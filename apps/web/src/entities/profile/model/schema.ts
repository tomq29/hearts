import { z } from "zod";

export const profileSchema = z.object({
  firstName: z.string().min(2, "Name is too short"),
  bio: z.string().max(500, "Bio is too long").optional(),
  gender: z.enum(["male", "female", "other"]),
  height: z.number().min(100).max(250).optional(),
  birthDate: z
    .string()
    .refine((date) => new Date(date).toString() !== "Invalid Date", {
      message: "Invalid date",
    }),
  photos: z.array(z.string()).min(1, "At least one photo is required"),
});

export type ProfileFormSchema = z.infer<typeof profileSchema>;
