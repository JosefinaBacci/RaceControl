import { StyleSheet, View } from 'react-native';

import type { HistoryType } from '@/api/history';
import { ChipGroup } from '@/components';
import { spacing } from '@/theme';

import { DateRangeFilter } from './DateRangeFilter';
import { allTypes, type DateRangeErrors, type HistoryFilterValue } from './historyFilter';

type HistoryFiltersProps = {
  value: HistoryFilterValue;
  onChange: (value: HistoryFilterValue) => void;
  errors: DateRangeErrors;
  types: readonly HistoryType[];
  typeLabel: string;
};

export function HistoryFilters({ value, onChange, errors, types, typeLabel }: HistoryFiltersProps) {
  const typeOptions = [{ value: allTypes, label: 'Todos' }, ...types.map((type) => ({ value: type.code, label: type.name }))];

  return (
    <View style={styles.filters}>
      <DateRangeFilter value={value} onChange={onChange} errors={errors} />
      <ChipGroup
        options={typeOptions}
        selected={value.type}
        onSelect={(type) => onChange({ ...value, type })}
        accessibilityLabel={`Filtrar por ${typeLabel}`}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  filters: { gap: spacing.md },
});
