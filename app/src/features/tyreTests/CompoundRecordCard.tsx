import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { Button, Card } from '@/components';
import { ResultChain } from '@/features/history/ResultChain';
import { spacing } from '@/theme';

import { TyreCorrectionForm } from './TyreCorrectionForm';
import type { CompoundRecord } from './tyreRecord';

export function CompoundRecordCard({ record, onCorrected }: { record: CompoundRecord; onCorrected: () => void }) {
  const [isCorrecting, setIsCorrecting] = useState(false);

  const corrected = () => {
    setIsCorrecting(false);
    onCorrected();
  };

  return (
    <Card title={record.compound.name}>
      <ResultChain revisions={record.revisions} />
      {isCorrecting ? (
        <TyreCorrectionForm current={record.current} onSaved={corrected} onDismiss={() => setIsCorrecting(false)} />
      ) : (
        <View style={styles.actions}>
          <Button label="Corregir" icon="create-outline" variant="secondary" compact onPress={() => setIsCorrecting(true)} />
        </View>
      )}
    </Card>
  );
}

const styles = StyleSheet.create({
  actions: { flexDirection: 'row', justifyContent: 'flex-end', gap: spacing.sm },
});
