import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { ControlResult } from '@/api/technicalControls';
import { ChipGroup } from '@/components';
import { useControlHistory, useControlTypes } from '@/data/technicalControls';
import { categoryFilterOptions } from '@/features/filters';
import { HistoryFilters } from '@/features/history/HistoryFilters';
import { HistoryList } from '@/features/history/HistoryList';
import { useHistoryFilter } from '@/features/history/useHistoryFilter';
import { RemoteContent } from '@/features/RemoteContent';
import type { CategoryFilter } from '@/mocks/catalog';
import { spacing } from '@/theme';

const subjectOf = (result: ControlResult) => result.team.name;

// The public history of technical controls (US22): every team's result in force,
// read-only, for the public and the teams alike.
export function ControlHistory() {
  const filter = useHistoryFilter();
  const [category, setCategory] = useState<CategoryFilter>('all');
  const types = useControlTypes();
  const { state, reload } = useControlHistory({ ...filter.query, category: category === 'all' ? undefined : category });

  return (
    <View style={styles.section}>
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <HistoryFilters
        value={filter.value}
        onChange={filter.setValue}
        errors={filter.errors}
        types={types.state.status === 'ready' ? types.state.data : []}
        typeLabel="tipo de control"
      />
      <RemoteContent state={state} onRetry={reload}>
        {(results) => (
          <HistoryList entries={results} subjectOf={subjectOf} emptyMessage="No hay controles técnicos para estos filtros." />
        )}
      </RemoteContent>
    </View>
  );
}

const styles = StyleSheet.create({
  section: { gap: spacing.md },
});
