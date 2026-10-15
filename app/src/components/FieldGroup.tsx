import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { spacing } from '@/theme';

import { AppText } from './AppText';

type FieldGroupProps = {
  label: string;
  error?: string;
  children: ReactNode;
};

// A labelled block inside a form: the overline names the field and the error, when
// present, sits right below whatever control the group wraps (a chip row, a picker).
export function FieldGroup({ label, error, children }: FieldGroupProps) {
  return (
    <View style={styles.group}>
      <AppText variant="overline" color="textMuted">
        {label}
      </AppText>
      {children}
      {error ? (
        <AppText variant="caption" color="danger" accessibilityLiveRegion="polite">
          {error}
        </AppText>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  group: { gap: spacing.sm },
});
