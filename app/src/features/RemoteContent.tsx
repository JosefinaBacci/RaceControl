import type { ReactNode } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import { Button, Notice } from '@/components';
import type { RemoteState } from '@/data/useRemote';
import { colors, spacing } from '@/theme';

type RemoteContentProps<T> = {
  state: RemoteState<T>;
  onRetry: () => void;
  children: (data: T) => ReactNode;
};

export function RemoteContent<T>({ state, onRetry, children }: RemoteContentProps<T>) {
  if (state.status === 'loading') {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.accent} />
      </View>
    );
  }
  if (state.status === 'error') {
    return (
      <View style={styles.error}>
        <Notice tone="danger" message={state.message} />
        <Button label="Reintentar" icon="refresh" variant="secondary" compact onPress={onRetry} />
      </View>
    );
  }
  return <>{children(state.data)}</>;
}

const styles = StyleSheet.create({
  center: { padding: spacing.xl, alignItems: 'center' },
  error: { gap: spacing.md, alignItems: 'flex-start' },
});
