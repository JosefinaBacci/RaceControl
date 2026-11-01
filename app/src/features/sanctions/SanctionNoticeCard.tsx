import type { SanctionNotice } from '@/api/sanctions';
import { AppText, Badge, Card, Notice } from '@/components';
import { useTeamSanctionNotices } from '@/data/TeamSanctionNoticesProvider';
import { AcknowledgementFooter } from '@/features/AcknowledgementFooter';
import { formatDateTime } from '@/format';

import { actionLabels, noticeAudienceText, penaltyText, subjectText } from './sanctionText';

// A sanction as a team of the championship receives it. The ones on the team or
// its drivers are flagged, since those concern it directly.
export function SanctionNoticeCard({ notice }: { notice: SanctionNotice }) {
  const { acknowledge } = useTeamSanctionNotices();
  const { sanction } = notice;
  const { current } = sanction;
  const action = actionLabels[current.action];
  const audience = noticeAudienceText(sanction.subject, notice.isOwnTeam);
  const context = [`${action.verb} el ${formatDateTime(current.changedAt)}`, current.event?.name, sanction.categoryName];

  return (
    <Card title={subjectText(sanction.subject, sanction.team)} action={<Badge label={action.label} tone={action.tone} />}>
      {audience ? <Notice tone="warning" message={audience} /> : null}
      <AppText variant="bodyStrong">{penaltyText(current.penalty)}</AppText>
      <AppText>{current.reason}</AppText>
      {current.note ? <AppText color="textMuted">{`Motivo del cambio: ${current.note}`}</AppText> : null}
      <AppText variant="caption" color="textMuted">
        {context.filter(Boolean).join(' · ')}
      </AppText>
      <AcknowledgementFooter acknowledgement={notice} onConfirm={() => acknowledge(notice.id)} />
    </Card>
  );
}
