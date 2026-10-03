import type { ReactNode } from 'react';
import { ScrollView, StyleSheet, View } from 'react-native';

import { colors, layout, spacing } from '@/theme';

import { AppText } from './AppText';

type ScreenProps = {
  title?: string;
  subtitle?: string;
  headerAction?: ReactNode;
  children: ReactNode;
};

export function Screen({ title, subtitle, headerAction, children }: ScreenProps) {
  return (
    <ScrollView style={styles.scroll} contentContainerStyle={styles.content} keyboardShouldPersistTaps="handled">
      <View style={styles.column}>
        {title ? (
          <View style={styles.header}>
            <View style={styles.headerText}>
              <AppText variant="title" accessibilityRole="header">
                {title}
              </AppText>
              {subtitle ? <AppText color="textMuted">{subtitle}</AppText> : null}
            </View>
            {headerAction}
          </View>
        ) : null}
        {children}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  scroll: { flex: 1, backgroundColor: colors.background },
  content: { padding: spacing.lg, paddingBottom: spacing.xxl, alignItems: 'center' },
  column: { width: '100%', maxWidth: layout.maxContentWidth, gap: spacing.lg },
  header: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: spacing.md, flexWrap: 'wrap' },
  headerText: { flexShrink: 1, gap: spacing.xs },
});
