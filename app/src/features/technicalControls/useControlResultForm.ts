import { useState } from 'react';

import type { ResultCorrectionInput, ResultInput } from '@/api/history';
import type { ControlRevision } from '@/api/technicalControls';
import { useFormSubmission } from '@/data/useFormSubmission';
import {
  correctionNoteError,
  emptyResultForm,
  resultPayload,
  validateResultForm,
  type ResultFormErrors,
} from '@/features/history/resultForm';

const fields = ['conductedOn', 'passed', 'notes', 'teamId', 'correctionNote'] as const;

type ControlResultErrors = ResultFormErrors & { correctionNote?: string };

const hasErrors = (errors: ControlResultErrors) => Object.values(errors).some(Boolean);

// Holds one team's result form: the shared fields plus, for a correction, its
// reason. The rules run here first and the API's answer lands on the same fields.
export function useControlResultForm(defaultDay: string) {
  const [values, setValues] = useState(() => emptyResultForm(defaultDay));
  const [correctionNote, setCorrectionNote] = useState('');
  const [localErrors, setLocalErrors] = useState<ControlResultErrors>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);

  const record = (action: (input: ResultInput) => Promise<ControlRevision>) => {
    const errors = validateResultForm(values);
    setLocalErrors(errors);
    return hasErrors(errors) ? Promise.resolve(null) : submit(() => action(resultPayload(values)));
  };

  const correct = (action: (input: ResultCorrectionInput) => Promise<ControlRevision>) => {
    const errors: ControlResultErrors = { ...validateResultForm(values), correctionNote: correctionNoteError(correctionNote) };
    setLocalErrors(errors);
    if (hasErrors(errors)) {
      return Promise.resolve(null);
    }
    return submit(() => action({ ...resultPayload(values), correctionNote: correctionNote.trim() }));
  };

  return {
    values,
    setValues,
    correctionNote,
    setCorrectionNote,
    errors: { ...localErrors, ...fieldErrors },
    formError: formError ?? fieldErrors.teamId ?? null,
    isSubmitting,
    record,
    correct,
  };
}
