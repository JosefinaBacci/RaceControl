import { ApiError } from './client';

export type FieldErrors<Field extends string> = Partial<Record<Field, string>>;

export type FormFailure<Field extends string> = {
  fieldErrors: FieldErrors<Field>;
  message: string | null;
};

export function describeFormFailure<Field extends string>(error: unknown, fields: readonly Field[]): FormFailure<Field> {
  if (!(error instanceof ApiError)) {
    return { fieldErrors: {}, message: 'Ocurrió un error inesperado. Intentá de nuevo.' };
  }
  const field = fields.find((candidate) => candidate === error.field);
  if (field) {
    return { fieldErrors: { [field]: error.message } as FieldErrors<Field>, message: null };
  }
  return { fieldErrors: {}, message: error.message };
}

export function hasFieldErrors<Field extends string>(errors: FieldErrors<Field>): boolean {
  return Object.values(errors).some(Boolean);
}
