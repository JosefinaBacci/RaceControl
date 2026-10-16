import { Badge } from '@/components';

type AcknowledgementBadgeProps = {
  acknowledged: boolean;
  acknowledgedLabel?: string;
  pendingLabel?: string;
};

export function AcknowledgementBadge({ acknowledged, acknowledgedLabel = 'Acusado', pendingLabel = 'Sin acuse' }: AcknowledgementBadgeProps) {
  if (acknowledged) {
    return <Badge label={acknowledgedLabel} tone="success" icon="checkmark" />;
  }
  return <Badge label={pendingLabel} tone="warning" icon="time-outline" />;
}
