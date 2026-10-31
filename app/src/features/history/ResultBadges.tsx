import { StyleSheet, View } from 'react-native';

import { Badge } from '@/components';
import { spacing } from '@/theme';

type ResultBadgesProps = {
  passed: boolean;
  isCorrection: boolean;
};

export function ResultBadges({ passed, isCorrection }: ResultBadgesProps) {
  return (
    <View style={styles.row}>
      <Badge
        label={passed ? 'Aprobado' : 'No aprobado'}
        tone={passed ? 'success' : 'danger'}
        icon={passed ? 'checkmark-circle' : 'close-circle'}
      />
      {isCorrection ? <Badge label="Corregido" tone="warning" icon="create-outline" /> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.xs },
});
