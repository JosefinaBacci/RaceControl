import { View } from 'react-native';

import type { ScoreNotice } from '@/api/scores';
import { AppText, Badge, Card, ListItem } from '@/components';
import { useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { AcknowledgementFooter } from '@/features/AcknowledgementFooter';
import { formatDateTime } from '@/format';

import { noticeKinds, pointsChangeText } from './scoreNotice';

export function ScoreNoticeCard({ notice }: { notice: ScoreNotice }) {
  const { acknowledge } = useTeamScoreNotices();
  const kind = noticeKinds[notice.kind];

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
      <AcknowledgementFooter acknowledgement={notice} onConfirm={() => acknowledge(notice.id)} />
    </Card>
  );
}
