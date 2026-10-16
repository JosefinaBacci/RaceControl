import { StyleSheet, View } from 'react-native';

import type { ScoreRevision } from '@/api/scores';
import { AppText, Badge, Card, ListItem } from '@/components';
import { formatDateTime } from '@/format';
import { spacing } from '@/theme';

import { noticeKinds, pointsChangeText } from './scoreNotice';

export function ScoreHistory({ revisions }: { revisions: ScoreRevision[] }) {
  return (
    <Card title="Historial de cambios">
      {revisions.length === 0 ? (
        <AppText color="textMuted">Todavía no se cargaron puntajes para esta carrera.</AppText>
      ) : (
        revisions.map((revision) => <RevisionEntry key={revision.id} revision={revision} />)
      )}
    </Card>
  );
}

function RevisionEntry({ revision }: { revision: ScoreRevision }) {
  const kind = noticeKinds[revision.kind];
  return (
    <View style={styles.revision}>
      <View style={styles.header}>
        <Badge label={kind.label} tone={kind.tone} />
        <AppText variant="caption" color="textMuted">
          {`${kind.verb} el ${formatDateTime(revision.changedAt)} por ${revision.changedBy}`}
        </AppText>
      </View>
      {revision.items.map((item, index) => (
        <ListItem
          key={item.driverId}
          title={`${item.firstName} ${item.lastName}`}
          trailing={<AppText variant="bodyStrong">{pointsChangeText(item)}</AppText>}
          isLast={index === revision.items.length - 1}
        />
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  revision: { gap: spacing.xs },
  header: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: spacing.sm },
});
