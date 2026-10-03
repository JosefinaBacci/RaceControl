import { useState } from 'react';

import { describeFormFailure, type FieldErrors } from '@/api/formFailure';

export function useFormSubmission<Field extends string>(fields: readonly Field[]) {
  const [fieldErrors, setFieldErrors] = useState<FieldErrors<Field>>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const submit = async <T,>(action: () => Promise<T>): Promise<T | null> => {
    setFieldErrors({});
    setFormError(null);
    setIsSubmitting(true);
    try {
      return await action();
    } catch (error) {
      const failure = describeFormFailure(error, fields);
      setFieldErrors(failure.fieldErrors);
      setFormError(failure.message);
      return null;
    } finally {
      setIsSubmitting(false);
    }
  };

  return { fieldErrors, formError, isSubmitting, submit, setFormError };
}
