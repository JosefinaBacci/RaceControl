import { StyleSheet, View } from 'react-native';

import type { TyreTestResult } from '@/api/tyreTests';
import { useTyreCompounds, useTyreTestHistory } from '@/data/tyreTests';
import { HistoryFilters } from '@/features/history/HistoryFilters';
import { HistoryList } from '@/features/history/HistoryList';
import { useHistoryFilter } from '@/features/history/useHistoryFilter';
import { RemoteContent } from '@/features/RemoteContent';
import { spacing } from '@/theme';

const subjectOf = (result: TyreTestResult) => `Prueba en ${result.circuitName}`;

// The public, read-only record of tyre tests: the current result of each compound
// tested, filtered by day and compound.
export function TyreTestHistory() {
  const filter = useHistoryFilter();
  const compounds = useTyreCompounds();
  const history = useTyreTestHistory(filter.query);

  return (
    <View style={styles.section}>
      <RemoteContent state={compounds.state} onRetry={compounds.reload}>
        {(types) => (
          <HistoryFilters value={filter.value} onChange={filter.setValue} errors={filter.errors} types={types} typeLabel="compuesto" />
        )}
      </RemoteContent>
      <RemoteContent state={history.state} onRetry={history.reload}>
        {(results) => (
          <HistoryList entries={results} subjectOf={subjectOf} emptyMessage="No hay pruebas de neumáticos para estos filtros." />
        )}
      </RemoteContent>
    </View>
  );
}

const styles = StyleSheet.create({
  section: { gap: spacing.md },
});
