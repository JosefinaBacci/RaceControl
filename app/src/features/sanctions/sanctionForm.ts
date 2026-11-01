import type { NewSanction, PenaltyType, Sanction, SanctionAmendment, SanctionTerms } from '@/api/sanctions';

export type SubjectKind = 'driver' | 'team';

export type SanctionFormValues = {
  subjectKind: SubjectKind;
  driverId: number | null;
  teamId: number | null;
  penaltyType: string;
  penaltyValue: string;
  reason: string;
  eventId: number | null;
  note: string;
};

// The names match the field the API reports on a rejection, so a 400 lands on
// the input that caused it.
export type SanctionFormField = 'target' | 'driverId' | 'teamId' | 'penaltyType' | 'penaltyValue' | 'reason' | 'eventId' | 'note';

export const sanctionFormFields: readonly SanctionFormField[] = [
  'target',
  'driverId',
  'teamId',
  'penaltyType',
  'penaltyValue',
  'reason',
  'eventId',
  'note',
];

export type SanctionFormErrors = Partial<Record<SanctionFormField, string>>;

// The same limits as the API (sanction_versions in 000035).
const minReasonLength = 3;
const maxReasonLength = 2000;
const maxNoteLength = 1000;

export const emptySanctionForm: SanctionFormValues = {
  subjectKind: 'driver',
  driverId: null,
  teamId: null,
  penaltyType: '',
  penaltyValue: '',
  reason: '',
  eventId: null,
  note: '',
};

export function formValuesOf(sanction: Sanction): SanctionFormValues {
  const { penalty, reason, event } = sanction.current;
  return {
    subjectKind: sanction.subject.kind,
    driverId: sanction.subject.driverId,
    teamId: sanction.subject.kind === 'team' ? sanction.team.id : null,
    penaltyType: penalty.code,
    penaltyValue: penalty.value === null ? '' : String(penalty.value),
    reason,
    eventId: event?.id ?? null,
    note: '',
  };
}

// The championship of the chosen subject, which limits the events a sanction
// can be tied to; null until a subject is chosen.
export function subjectCategory(
  values: SanctionFormValues,
  drivers: readonly { id: number; categoryCode: string }[],
  teams: readonly { id: number; categoryCode: string }[],
): string | null {
  const candidates = values.subjectKind === 'driver' ? drivers : teams;
  const chosenId = values.subjectKind === 'driver' ? values.driverId : values.teamId;
  return candidates.find((candidate) => candidate.id === chosenId)?.categoryCode ?? null;
}

export function penaltyTypeOf(values: SanctionFormValues, types: readonly PenaltyType[]): PenaltyType | undefined {
  return types.find((type) => type.code === values.penaltyType);
}

export function validateTerms(values: SanctionFormValues, types: readonly PenaltyType[]): SanctionFormErrors {
  const errors: SanctionFormErrors = {};
  const reasonLength = values.reason.trim().length;
  if (reasonLength < minReasonLength || reasonLength > maxReasonLength) {
    errors.reason = `El motivo es obligatorio (entre ${minReasonLength} y ${maxReasonLength} caracteres)`;
  }
  const penaltyType = penaltyTypeOf(values, types);
  if (!penaltyType) {
    errors.penaltyType = 'Elegí la penalización';
  } else if (penaltyType.maxValue !== null && !isValueWithin(values.penaltyValue, penaltyType.maxValue)) {
    errors.penaltyValue = `Ingresá un entero entre 1 y ${penaltyType.maxValue}`;
  }
  return errors;
}

export function validateNewSanction(values: SanctionFormValues, types: readonly PenaltyType[]): SanctionFormErrors {
  const errors = validateTerms(values, types);
  if (values.subjectKind === 'driver' && values.driverId === null) {
    errors.driverId = 'Elegí el piloto sancionado';
  }
  if (values.subjectKind === 'team' && values.teamId === null) {
    errors.teamId = 'Elegí la escudería sancionada';
  }
  return errors;
}

export function validateAmendment(values: SanctionFormValues, types: readonly PenaltyType[]): SanctionFormErrors {
  return { ...validateTerms(values, types), ...noteError(values.note) };
}

export function noteError(note: string): SanctionFormErrors {
  const length = note.trim().length;
  if (length === 0 || length > maxNoteLength) {
    return { note: `Explicá el motivo del cambio (hasta ${maxNoteLength} caracteres)` };
  }
  return {};
}

function isValueWithin(raw: string, maxValue: number): boolean {
  if (!/^\d+$/.test(raw.trim())) {
    return false;
  }
  const value = Number(raw);
  return value >= 1 && value <= maxValue;
}

function termsOf(values: SanctionFormValues, types: readonly PenaltyType[]): SanctionTerms {
  const takesValue = penaltyTypeOf(values, types)?.maxValue != null;
  return {
    penaltyType: values.penaltyType,
    penaltyValue: takesValue ? Number(values.penaltyValue) : null,
    reason: values.reason.trim(),
    eventId: values.eventId,
  };
}

// Only the chosen subject travels: the API refuses a sanction with both.
export function toNewSanction(values: SanctionFormValues, types: readonly PenaltyType[]): NewSanction {
  const subject = values.subjectKind === 'driver' ? { driverId: values.driverId ?? undefined } : { teamId: values.teamId ?? undefined };
  return { ...subject, ...termsOf(values, types) };
}

export function toAmendment(values: SanctionFormValues, types: readonly PenaltyType[]): SanctionAmendment {
  return { ...termsOf(values, types), note: values.note.trim() };
}
