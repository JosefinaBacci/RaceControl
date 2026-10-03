import Ionicons from '@expo/vector-icons/Ionicons';
import type { ComponentProps } from 'react';
import { StyleSheet, View } from 'react-native';

import { colors, radius, spacing, type ColorName } from '@/theme';

import { AppText } from './AppText';

export type BadgeTone = 'neutral' | 'info' | 'success' | 'warning' | 'danger' | 'accent';

const tones: Record<BadgeTone, { background: ColorName; text: ColorName }> = {
  neutral: { background: 'surfaceMuted', text: 'textMuted' },
  info: { background: 'infoSoft', text: 'info' },
  success: { background: 'successSoft', text: 'success' },
  warning: { background: 'warningSoft', text: 'warning' },
  danger: { background: 'dangerSoft', text: 'danger' },
  accent: { background: 'accentSoft', text: 'accent' },
};

type BadgeProps = {
  label: string;
  tone?: BadgeTone;
  icon?: ComponentProps<typeof Ionicons>['name'];
};

export function Badge({ label, tone = 'neutral', icon }: BadgeProps) {
  const palette = tones[tone];
  return (
    <View style={[styles.badge, { backgroundColor: colors[palette.background] }]}>
      {icon ? <Ionicons name={icon} size={12} color={colors[palette.text]} /> : null}
      <AppText variant="caption" color={palette.text}>
        {label}
      </AppText>
    </View>
  );
}

const styles = StyleSheet.create({
  badge: {
    alignSelf: 'flex-start',
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: spacing.sm + 2,
    paddingVertical: 3,
    borderRadius: radius.sm,
  },
});
