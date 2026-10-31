import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { hasFieldErrors } from '@/api/formFailure';
import { amendSanction, revokeSanction, type PenaltyType, type Sanction } from '@/api/sanctions';
import { Button, Card, Notice, TextField } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { spacing } from '@/theme';

import { TermsFields } from './SanctionFields';
import {
  formValuesOf,
  noteError,
  sanctionFormFields,
  toAmendment,
  validateAmendment,
  type SanctionFormErrors,
} from './sanctionForm';

type AdminFormProps = {
  sanction: Sanction;
  onSaved: () => void;
};

// An amendment replaces the terms and asks every team to confirm again; the
// subject cannot change, a wrong one is revoked and registered anew.
export function AmendSanctionCard({ sanction, penaltyTypes, onSaved }: AdminFormProps & { penaltyTypes: readonly PenaltyType[] }) {
  const [values, setValues] = useState(() => formValuesOf(sanction));
  const [localErrors, setLocalErrors] = useState<SanctionFormErrors>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(sanctionFormFields);

  const save = async () => {
    const errors = validateAmendment(values, penaltyTypes);
    setLocalErrors(errors);
    if (hasFieldErrors(errors)) {
      return;
    }
    if (await submit(() => amendSanction(sanction.id, toAmendment(values, penaltyTypes)))) {
      setValues({ ...values, note: '' });
      onSaved();
    }
  };

  const errors = { ...fieldErrors, ...localErrors };
  return (
    <Card title="Modificar sanción">
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <TermsFields values={values} onChange={setValues} errors={errors} penaltyTypes={penaltyTypes} categoryCode={sanction.categoryCode} />
      <TextField label="Motivo del cambio" value={values.note} onChangeText={(note) => setValues({ ...values, note })} error={errors.note} multiline />
      <View style={styles.actions}>
        <Button label="Guardar modificación" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

export function RevokeSanctionCard({ sanction, onSaved }: AdminFormProps) {
  const [note, setNote] = useState('');
  const [localErrors, setLocalErrors] = useState<SanctionFormErrors>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(sanctionFormFields);

  const revoke = async () => {
    const errors = noteError(note);
    setLocalErrors(errors);
    if (hasFieldErrors(errors)) {
      return;
    }
    if (await submit(() => revokeSanction(sanction.id, note.trim()))) {
      onSaved();
    }
  };

  return (
    <Card title="Revocar sanción">
      <Notice tone="warning" message="La revocación queda en el registro público con su motivo y se notifica a las escuderías." />
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <TextField label="Motivo de la revocación" value={note} onChangeText={setNote} error={localErrors.note ?? fieldErrors.note} multiline />
      <View style={styles.actions}>
        <Button label="Revocar" icon="close-circle-outline" variant="secondary" onPress={revoke} loading={isSubmitting} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.md },
});
