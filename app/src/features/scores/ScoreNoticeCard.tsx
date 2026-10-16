import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { ScoreNotice } from '@/api/scores';
import { AppText, Badge, Button, Card, ListItem, Notice } from '@/components';
import { useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { formatDateTime } from '@/format';
import { spacing } from '@/theme';

import { acknowledgementText, noticeKinds, pointsChangeText } from './scoreNotice';

export function ScoreNoticeCard({ notice }: { notice: ScoreNotice }) {
  const kind = noticeKinds[notice.kind];
  const isConfirmed = notice.status === 'confirmed';

  return (
    <Card title={notice.race.name} action={<Badge label={kind.label} tone={kind.tone} />}>
      <AppText variant="caption" color="textMuted">
        {`${kind.verb} el ${formatDateTime(notice.changedAt)} · ${notice.race.categoryName}`}
      </AppText>
      <View>
        {notice.items.map((item, index) => (
          <ListItem
            key={item.driverId}
            title={`${item.firstName} ${item.lastName}`}
            trailing={<AppText variant="bodyStrong">{pointsChangeText(item)}</AppText>}
            isLast={index === notice.items.length - 1}
          />
        ))}
      </View>
      <View style={styles.footer}>
        <View style={styles.status}>
          <AcknowledgementBadge acknowledged={isConfirmed} acknowledgedLabel="Confirmada" pendingLabel="Pendiente" />
          <AppText variant="caption" color="textMuted">
            {acknowledgementText(notice)}
          </AppText>
        </View>
        {isConfirmed ? null : <ConfirmButton changeLogId={notice.id} />}
      </View>
    </Card>
  );
}

function ConfirmButton({ changeLogId }: { changeLogId: number }) {
  const { acknowledge } = useTeamScoreNotices();
  const [isConfirming, setIsConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const confirm = async () => {
    setError(null);
    setIsConfirming(true);
    try {
      await acknowledge(changeLogId);
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'No se pudo confirmar la recepción.');
    } finally {
      setIsConfirming(false);
    }
  };

  return (
    <View style={styles.confirm}>
      {error ? <Notice tone="danger" message={error} /> : null}
      <Button label="Confirmar recepción" icon="checkmark-done" compact onPress={confirm} loading={isConfirming} />
    </View>
  );
}

const styles = StyleSheet.create({
  footer: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', justifyContent: 'space-between', gap: spacing.md },
  status: { gap: spacing.xs, flexShrink: 1 },
  confirm: { gap: spacing.sm, alignItems: 'flex-end' },
});
