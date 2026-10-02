import { ApiError } from '@/api/client';

export type LoginFieldErrors = {
  username?: string;
  password?: string;
};

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

export function hasErrors(errors: LoginFieldErrors): boolean {
  return Object.values(errors).some(Boolean);
}

export type LoginFailure = { fieldErrors: LoginFieldErrors; message: string | null };

export function describeLoginFailure(error: unknown): LoginFailure {
  if (!(error instanceof ApiError)) {
    return { fieldErrors: {}, message: 'Ocurrió un error inesperado. Intentá de nuevo.' };
  }
  if (error.code === 'validation' && (error.field === 'username' || error.field === 'password')) {
    return { fieldErrors: { [error.field]: error.message }, message: null };
  }
  return { fieldErrors: {}, message: error.message };
}
