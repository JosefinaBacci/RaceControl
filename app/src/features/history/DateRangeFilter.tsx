import { StyleSheet, View } from 'react-native';

import { TextField } from '@/components';
import { spacing } from '@/theme';

import type { DateRangeErrors, DateRangeValue } from './historyFilter';

type DateRangeFilterProps<T extends DateRangeValue> = {
  value: T;
  onChange: (value: T) => void;
  errors?: DateRangeErrors;
};

// Generic over the full filter value so a screen with more filters (type,
// team) keeps them while the range changes.
export function DateRangeFilter<T extends DateRangeValue>({ value, onChange, errors = {} }: DateRangeFilterProps<T>) {
  return (
    <View style={styles.row}>
      <View style={styles.field}>
        <TextField
          label="Desde (AAAA-MM-DD)"
          icon="calendar-outline"
          value={value.from}
          onChangeText={(from) => onChange({ ...value, from })}
          error={errors.from}
          autoCorrect={false}
          inputMode="numeric"
        />
      </View>
      <View style={styles.field}>
        <TextField
          label="Hasta (AAAA-MM-DD)"
          icon="calendar-outline"
          value={value.to}
          onChangeText={(to) => onChange({ ...value, to })}
          error={errors.to}
          autoCorrect={false}
          inputMode="numeric"
        />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  field: { flexGrow: 1, flexBasis: 160 },
});
