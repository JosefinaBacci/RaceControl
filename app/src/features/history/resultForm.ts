import { isDay } from './day';

// The form block every official result shares: the day it was obtained, whether
// it passed and the notes. Mirrors the API rules (internal/results) so the FIA
// sees the mistake before sending it.

export const maxNotesLength = 2000;

export type ResultFormValues = {
  conductedOn: string;
  passed: boolean | null;
  notes: string;
};

export type ResultField = keyof ResultFormValues;

export type ResultFormErrors = Partial<Record<ResultField, string>>;

export const resultFields: readonly ResultField[] = ['conductedOn', 'passed', 'notes'];

export function emptyResultForm(today: string): ResultFormValues {
  return { conductedOn: today, passed: null, notes: '' };
}

export function validateResultForm(values: ResultFormValues): ResultFormErrors {
  const errors: ResultFormErrors = {};
  const notes = values.notes.trim();
  if (!isDay(values.conductedOn.trim())) {
    errors.conductedOn = 'Usá el formato AAAA-MM-DD';
  }
  if (values.passed === null) {
    errors.passed = 'Indicá si aprobó o no';
  }
  if (values.passed === false && !notes) {
    errors.notes = 'Un resultado no aprobado necesita observaciones';
  }
  if (notes.length > maxNotesLength) {
    errors.notes = 'Las observaciones no pueden superar los 2000 caracteres';
  }
  return errors;
}

export function correctionNoteError(note: string): string | undefined {
  const trimmed = note.trim();
  if (!trimmed) {
    return 'La corrección necesita un motivo';
  }
  return trimmed.length > maxNotesLength ? 'El motivo no puede superar los 2000 caracteres' : undefined;
}

// Only called once validateResultForm found no errors, so passed is set.
export function resultPayload(values: ResultFormValues): { conductedOn: string; passed: boolean; notes: string } {
  return { conductedOn: values.conductedOn.trim(), passed: values.passed === true, notes: values.notes.trim() };
}
