import { z } from 'zod';

const PASSWORD_REGEX =
  /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/;

export const signInSchema = z.object({
  email: z.string().trim().min(1, 'Email is required').email('Invalid email'),
  password: z.string().min(1, 'Password is required'),
});

export const signUpSchema = z
  .object({
    first_name: z.string().min(1, 'First name is required').max(50),
    last_name: z.string().min(1, 'Last name is required').max(50),
    email: z
      .string()
      .trim()
      .min(1, 'Email is required')
      .email('Invalid email'),
    password: z
      .string()
      .regex(
        PASSWORD_REGEX,
        'Min 8 chars, uppercase, lowercase, number, and special char (@$!%*?&)',
      ),
    confirm_password: z.string().min(1, 'Please confirm your password'),
  })
  .refine((d) => d.password === d.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  });

export const changePasswordSchema = z
  .object({
    current_password: z.string().min(1, 'Current password is required'),
    new_password: z
      .string()
      .regex(
        PASSWORD_REGEX,
        'Min 8 chars, uppercase, lowercase, number, and special char (@$!%*?&)',
      ),
    confirm_password: z.string().min(1, 'Please confirm your password'),
  })
  .refine((d) => d.new_password === d.confirm_password, {
    message: 'Passwords do not match',
    path: ['confirm_password'],
  });

export const profileSchema = z.object({
  first_name: z.string().max(50).optional().or(z.literal('')),
  last_name: z.string().max(50).optional().or(z.literal('')),
  phone_number: z.string().min(7).optional().or(z.literal('')),
  country: z.string().max(50).optional().or(z.literal('')),
  date_of_birth: z.string().optional().or(z.literal('')),
  bio: z.string().max(500).optional().or(z.literal('')),
});
