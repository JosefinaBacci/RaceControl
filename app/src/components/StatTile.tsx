import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

type StatTileProps = {
  label: string;
  value: string | number;
  hint?: string;
  highlight?: boolean;
};

export function StatTile({ label, value, hint, highlight = false }: StatTileProps) {
  return (
    <View style={[styles.tile, highlight && styles.highlight]}>
      <AppText variant="display" color={highlight ? 'accent' : 'text'}>
        {value}
      </AppText>
      <AppText variant="overline" color="textMuted">
        {label}
      </AppText>
      {hint ? (
        <AppText variant="caption" color="textMuted">
          {hint}
        </AppText>
      ) : null}
    </View>
  );
}

export function StatRow({ children }: { children: ReactNode }) {
  return <View style={styles.row}>{children}</View>;
}

const styles = StyleSheet.create({
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.md },
  tile: {
    flexGrow: 1,
    flexBasis: 150,
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    padding: spacing.lg,
    gap: spacing.xs,
  },
  highlight: { borderColor: colors.accent, backgroundColor: colors.accentSoft },
});
