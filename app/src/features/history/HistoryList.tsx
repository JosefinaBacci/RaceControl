import { StyleSheet, View } from 'react-native';

import type { HistoryEntry } from '@/api/history';
import { AppText, Card, Notice } from '@/components';
import { spacing } from '@/theme';

import { formatDay } from './day';
import { ResultBadges } from './ResultBadges';

type HistoryListProps<T extends HistoryEntry> = {
  entries: readonly T[];
  subjectOf: (entry: T) => string;
  emptyMessage: string;
};

// Read-only by design: the public and the teams consult the official record, so
// no entry offers an action.
export function HistoryList<T extends HistoryEntry>({ entries, subjectOf, emptyMessage }: HistoryListProps<T>) {
  if (entries.length === 0) {
    return <Notice tone="info" message={emptyMessage} />;
  }
  return (
    <View style={styles.list}>
      {entries.map((entry) => (
        <HistoryEntryCard key={entry.id} entry={entry} subject={subjectOf(entry)} />
      ))}
    </View>
  );
}

function HistoryEntryCard({ entry, subject }: { entry: HistoryEntry; subject: string }) {
  return (
    <Card>
      <View style={styles.header}>
        <AppText variant="heading" style={styles.subject}>
          {subject}
        </AppText>
        <ResultBadges passed={entry.passed} isCorrection={entry.isCorrection} />
      </View>
      <AppText variant="caption" color="textMuted">
        {`${formatDay(entry.conductedOn)} · ${entry.type.name} · ${entry.categoryCode.toUpperCase()} ${entry.seasonYear}`}
      </AppText>
      {entry.notes ? <AppText>{entry.notes}</AppText> : null}
      {entry.correctionNote ? (
        <AppText variant="caption" color="textMuted">{`Corrección: ${entry.correctionNote}`}</AppText>
      ) : null}
    </Card>
  );
}

const styles = StyleSheet.create({
  list: { gap: spacing.md },
  // On a narrow phone the badges drop under the subject instead of being clipped.
  header: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: spacing.sm },
  subject: { flexShrink: 1 },
});
