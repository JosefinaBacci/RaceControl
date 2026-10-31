import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { Acknowledgement } from '@/api/scores';
import { AppText, Button, Notice } from '@/components';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { acknowledgementText } from '@/features/scores/scoreNotice';
import { spacing } from '@/theme';

type AcknowledgementFooterProps = {
  acknowledgement: Acknowledgement;
  onConfirm: () => Promise<void>;
};

// Where a team stands on a notice it was sent, and the button to tell the FIA it
// received it while it is still pending.
export function AcknowledgementFooter({ acknowledgement, onConfirm }: AcknowledgementFooterProps) {
  const isConfirmed = acknowledgement.status === 'confirmed';

  return (
    <View style={styles.footer}>
      <View style={styles.status}>
        <AcknowledgementBadge acknowledged={isConfirmed} acknowledgedLabel="Confirmada" pendingLabel="Pendiente" />
        <AppText variant="caption" color="textMuted">
          {acknowledgementText(acknowledgement)}
        </AppText>
      </View>
      {isConfirmed ? null : <ConfirmReceiptButton onConfirm={onConfirm} />}
    </View>
  );
}

function ConfirmReceiptButton({ onConfirm }: { onConfirm: () => Promise<void> }) {
  const [isConfirming, setIsConfirming] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const confirm = async () => {
    setError(null);
    setIsConfirming(true);
    try {
      await onConfirm();
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
