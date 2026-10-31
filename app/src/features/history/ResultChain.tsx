import { StyleSheet, View } from 'react-native';

import type { ResultRevision } from '@/api/history';
import { AppText, Badge } from '@/components';
import { formatDateTime } from '@/format';
import { colors, spacing } from '@/theme';

import { formatDay } from './day';
import { ResultBadges } from './ResultBadges';

// The FIA's view of one subject's record: every entry it ever had, newest first,
// with who recorded it. Only the newest one is in force.
export function ResultChain({ revisions }: { revisions: readonly ResultRevision[] }) {
  const newestFirst = [...revisions].sort((left, right) => right.id - left.id);
  return (
    <View style={styles.chain}>
      {newestFirst.map((revision, index) => (
        <View key={revision.id} style={[styles.revision, index > 0 && styles.superseded]}>
          <View style={styles.header}>
            <ResultBadges passed={revision.passed} isCorrection={revision.isCorrection} />
            {index === 0 ? <Badge label="Vigente" tone="info" /> : <Badge label="Reemplazado" />}
          </View>
          <AppText variant="caption" color="textMuted">
            {`Realizado el ${formatDay(revision.conductedOn)} · registrado por ${revision.recordedBy.username} el ${formatDateTime(revision.recordedAt)}`}
          </AppText>
          {revision.notes ? <AppText>{revision.notes}</AppText> : null}
          {revision.correctionNote ? (
            <AppText variant="caption" color="textMuted">{`Motivo de la corrección: ${revision.correctionNote}`}</AppText>
          ) : null}
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  chain: { gap: spacing.sm },
  revision: { gap: spacing.xs },
  superseded: { opacity: 0.7, borderTopWidth: 1, borderTopColor: colors.border, paddingTop: spacing.sm },
  header: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs },
});
