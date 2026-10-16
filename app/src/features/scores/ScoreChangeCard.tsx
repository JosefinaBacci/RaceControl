import type { ScoreChange } from '@/api/scores';
import { AppText, Avatar, Badge, Card, ListItem } from '@/components';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { formatDateTime } from '@/format';
import { teamLogo } from '@/logos';
import { teamById } from '@/mocks/catalog';

import { acknowledgementText, noticeKinds } from './scoreNotice';

// A score change as the FIA follows it: every team it touched and whether each
// one confirmed receiving it.
export function ScoreChangeCard({ change }: { change: ScoreChange }) {
  const kind = noticeKinds[change.kind];

  return (
    <Card title={change.race.name} action={<Badge label={kind.label} tone={kind.tone} />}>
      <AppText variant="caption" color="textMuted">
        {`${kind.verb} por ${change.changedBy} el ${formatDateTime(change.changedAt)} · ${change.race.categoryName}`}
      </AppText>
      {change.teams.map((team, index) => (
        <ListItem
          key={team.teamId}
          leading={<Avatar name={team.teamName} color={teamById(team.teamId).color} image={teamLogo(team.teamId)} />}
          title={team.teamName}
          subtitle={acknowledgementText(team)}
          trailing={<AcknowledgementBadge acknowledged={team.status === 'confirmed'} acknowledgedLabel="Confirmada" pendingLabel="Pendiente" />}
          isLast={index === change.teams.length - 1}
        />
      ))}
    </Card>
  );
}
