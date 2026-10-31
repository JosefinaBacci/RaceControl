import { StyleSheet, View } from 'react-native';

import { ChipGroup, FieldGroup, TextField } from '@/components';
import { spacing } from '@/theme';

import type { ResultFormErrors, ResultFormValues } from './resultForm';

type ResultFieldsProps = {
  values: ResultFormValues;
  onChange: (values: ResultFormValues) => void;
  errors: ResultFormErrors;
};

const unset = 'unset';

const outcomeOptions = [
  { value: 'passed' as const, label: 'Aprobado' },
  { value: 'failed' as const, label: 'No aprobado' },
];

type OutcomeChoice = (typeof outcomeOptions)[number]['value'] | typeof unset;

function choiceOf(passed: boolean | null): OutcomeChoice {
  if (passed === null) {
    return unset;
  }
  return passed ? 'passed' : 'failed';
}

// The fields every official result shares. The screen adds what the result
// grades (a team, a compound) and, for a correction, its reason.
export function ResultFields({ values, onChange, errors }: ResultFieldsProps) {
  return (
    <View style={styles.fields}>
      <TextField
        label="Fecha en que se realizó (AAAA-MM-DD)"
        icon="calendar-outline"
        value={values.conductedOn}
        onChangeText={(conductedOn) => onChange({ ...values, conductedOn })}
        error={errors.conductedOn}
        autoCorrect={false}
        inputMode="numeric"
      />
      <FieldGroup label="Resultado" error={errors.passed}>
        <ChipGroup<OutcomeChoice>
          options={outcomeOptions}
          selected={choiceOf(values.passed)}
          onSelect={(choice) => onChange({ ...values, passed: choice === 'passed' })}
          accessibilityLabel="Resultado"
        />
      </FieldGroup>
      <TextField
        label="Observaciones"
        value={values.notes}
        onChangeText={(notes) => onChange({ ...values, notes })}
        error={errors.notes}
        multiline
      />
    </View>
  );
}

const styles = StyleSheet.create({
  fields: { gap: spacing.md },
});
