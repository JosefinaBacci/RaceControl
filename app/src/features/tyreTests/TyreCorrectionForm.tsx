import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { correctTyreTestResult, type TyreTestRevision } from '@/api/tyreTests';
import { Button, Notice, TextField } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { ResultFields } from '@/features/history/ResultFields';
import {
  correctionNoteError,
  resultFields,
  resultPayload,
  validateResultForm,
  type ResultFormErrors,
  type ResultFormValues,
} from '@/features/history/resultForm';
import { spacing } from '@/theme';

const fields = [...resultFields, 'correctionNote'] as const;

type TyreCorrectionFormProps = {
  current: TyreTestRevision;
  onSaved: () => void;
  onDismiss: () => void;
};

function valuesOf(revision: TyreTestRevision): ResultFormValues {
  return { conductedOn: revision.conductedOn, passed: revision.passed, notes: revision.notes };
}

// A result is never edited: the correction is a new result that replaces the
// current one and keeps the reason on record.
export function TyreCorrectionForm({ current, onSaved, onDismiss }: TyreCorrectionFormProps) {
  const [values, setValues] = useState(() => valuesOf(current));
  const [correctionNote, setCorrectionNote] = useState('');
  const [localErrors, setLocalErrors] = useState<ResultFormErrors & { correctionNote?: string }>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);
  const errors = { ...fieldErrors, ...localErrors };

  const save = async () => {
    const found = { ...validateResultForm(values), correctionNote: correctionNoteError(correctionNote) };
    setLocalErrors(found);
    if (Object.values(found).some(Boolean)) {
      return;
    }
    const saved = await submit(() =>
      correctTyreTestResult(current.id, { ...resultPayload(values), correctionNote: correctionNote.trim() }),
    );
    if (saved) {
      onSaved();
    }
  };

  return (
    <View style={styles.form}>
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <ResultFields values={values} onChange={setValues} errors={errors} />
      <TextField
        label="Motivo de la corrección"
        value={correctionNote}
        onChangeText={setCorrectionNote}
        error={errors.correctionNote}
        multiline
      />
      <View style={styles.actions}>
        <Button label="Descartar" variant="ghost" onPress={onDismiss} />
        <Button label="Guardar corrección" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  form: { gap: spacing.md },
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.sm },
});
