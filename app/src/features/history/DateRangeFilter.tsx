import { StyleSheet, View } from 'react-native';

import { FieldGroup, TextField } from '@/components';
import { spacing } from '@/theme';

import type { DateRangeErrors, DateRangeValue } from './historyFilter';

type DateRangeFilterProps<T extends DateRangeValue> = {
  value: T;
  onChange: (value: T) => void;
  errors?: DateRangeErrors;
};

const dayHint = 'AAAA-MM-DD';

// Wide enough that two bounds only share a row when the hint fits whole: on a
// phone they stack instead of clipping it.
const minFieldWidth = 220;

// Generic over the full filter value so a screen with more filters (type,
// team) keeps them while the range changes.
export function DateRangeFilter<T extends DateRangeValue>({ value, onChange, errors = {} }: DateRangeFilterProps<T>) {
  return (
    <View style={styles.row}>
      <View style={styles.field}>
        <FieldGroup label="Desde">
          <TextField
            label="Desde"
            placeholder={dayHint}
            icon="calendar-outline"
            value={value.from}
            onChangeText={(from) => onChange({ ...value, from })}
            error={errors.from}
            autoCorrect={false}
            inputMode="numeric"
          />
        </FieldGroup>
      </View>
      <View style={styles.field}>
        <FieldGroup label="Hasta">
          <TextField
            label="Hasta"
            placeholder={dayHint}
            icon="calendar-outline"
            value={value.to}
            onChangeText={(to) => onChange({ ...value, to })}
            error={errors.to}
            autoCorrect={false}
            inputMode="numeric"
          />
        </FieldGroup>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  field: { flexGrow: 1, flexBasis: minFieldWidth },
});
