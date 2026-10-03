import { describeFormFailure, hasFieldErrors, type FieldErrors, type FormFailure } from '@/api/formFailure';

const loginFields = ['username', 'password'] as const;
type LoginField = (typeof loginFields)[number];

export type LoginFieldErrors = FieldErrors<LoginField>;

const usernamePattern = /^[a-z0-9._-]{3,32}$/;
const maxPasswordLength = 128;

export function normalizeUsername(raw: string): string {
  return raw.trim().toLowerCase();
}

export function validateLoginForm(username: string, password: string): LoginFieldErrors {
  const errors: LoginFieldErrors = {};

  if (username.length === 0) {
    errors.username = 'Ingresá tu usuario';
  } else if (!usernamePattern.test(username)) {
    errors.username = 'El usuario solo admite letras, números, punto, guion y guion bajo (3 a 32 caracteres)';
  }

  if (password.length === 0) {
    errors.password = 'Ingresá tu contraseña';
  } else if (password.length > maxPasswordLength) {
    errors.password = 'La contraseña no puede superar los 128 caracteres';
  }

  return errors;
}

export const hasErrors = hasFieldErrors<LoginField>;

export function describeLoginFailure(error: unknown): FormFailure<LoginField> {
  return describeFormFailure(error, loginFields);
}
