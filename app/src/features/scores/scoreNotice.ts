import type { Acknowledgement, ChangedPoints, NoticeKind } from '@/api/scores';
import type { BadgeTone } from '@/components';
import { formatDateTime, formatPoints } from '@/format';

export const noticeKinds: Record<NoticeKind, { label: string; tone: BadgeTone; verb: string }> = {
  published: { label: 'Publicación', tone: 'info', verb: 'Publicados' },
  corrected: { label: 'Corrección', tone: 'warning', verb: 'Corregidos' },
};

export function pointsChangeText({ oldPoints, newPoints }: ChangedPoints): string {
  const after = newPoints === null ? 'sin puntaje' : `${formatPoints(newPoints)} pts`;
  return oldPoints === null ? after : `${formatPoints(oldPoints)} → ${after}`;
}

export function acknowledgementText({ status, acknowledgedAt, acknowledgedBy }: Acknowledgement): string {
  if (status === 'pending' || acknowledgedAt === null) {
    return 'Pendiente de confirmación';
  }
  return `Confirmada por ${acknowledgedBy ?? '—'} el ${formatDateTime(acknowledgedAt)}`;
}
