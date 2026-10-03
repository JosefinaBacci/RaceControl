import { Badge } from '@/components';

type AcknowledgementBadgeProps = {
  acknowledged: boolean;
  acknowledgedLabel?: string;
};

export function AcknowledgementBadge({ acknowledged, acknowledgedLabel = 'Acusado' }: AcknowledgementBadgeProps) {
  if (acknowledged) {
    return <Badge label={acknowledgedLabel} tone="success" icon="checkmark" />;
  }
  return <Badge label="Sin acuse" tone="warning" icon="time-outline" />;
}
