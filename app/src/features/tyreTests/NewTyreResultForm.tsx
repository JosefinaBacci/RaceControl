import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { HistoryType } from '@/api/history';
import { recordTyreTestResult } from '@/api/tyreTests';
import { Button, Card, ChipGroup, FieldGroup, Notice } from '@/components';
import { useFormSubmission } from '@/data/useFormSubmission';
import { ResultFields } from '@/features/history/ResultFields';
import {
  emptyResultForm,
  resultFields,
  resultPayload,
  validateResultForm,
  type ResultFormErrors,
} from '@/features/history/resultForm';
import { spacing } from '@/theme';

import { compoundError } from './tyreRecord';

const fields = [...resultFields, 'compoundCode'] as const;

type NewTyreResultFormProps = {
  eventId: number;
  compounds: readonly HistoryType[];
  defaultDay: string;
  onSaved: () => void;
};

export function NewTyreResultForm({ eventId, compounds, defaultDay, onSaved }: NewTyreResultFormProps) {
  const [values, setValues] = useState(() => emptyResultForm(defaultDay));
  const [compoundCode, setCompoundCode] = useState('');
  const [localErrors, setLocalErrors] = useState<ResultFormErrors & { compoundCode?: string }>({});
  const { fieldErrors, formError, isSubmitting, submit } = useFormSubmission(fields);
  const errors = { ...fieldErrors, ...localErrors };

  const save = async () => {
    const found = { ...validateResultForm(values), compoundCode: compoundError(compoundCode) };
    setLocalErrors(found);
    if (Object.values(found).some(Boolean)) {
      return;
    }
    const saved = await submit(() => recordTyreTestResult(eventId, { ...resultPayload(values), compoundCode }));
    if (saved) {
      setValues(emptyResultForm(defaultDay));
      setCompoundCode('');
      onSaved();
    }
  };

  if (compounds.length === 0) {
    return <Notice tone="info" message="Todos los compuestos ya tienen resultado: corregilos desde su registro." />;
  }
  return (
    <Card title="Registrar resultado">
      {formError ? <Notice tone="danger" message={formError} /> : null}
      <FieldGroup label="Compuesto" error={errors.compoundCode}>
        <ChipGroup
          options={compounds.map((compound) => ({ value: compound.code, label: compound.name }))}
          selected={compoundCode}
          onSelect={setCompoundCode}
          accessibilityLabel="Compuesto"
        />
      </FieldGroup>
      <ResultFields values={values} onChange={setValues} errors={errors} />
      <View style={styles.actions}>
        <Button label="Registrar resultado" icon="save-outline" onPress={save} loading={isSubmitting} />
      </View>
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.sm },
});
