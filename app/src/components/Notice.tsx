import Ionicons from '@expo/vector-icons/Ionicons';
import { StyleSheet, View } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

const tones = {
  info: { background: colors.infoSoft, text: 'info', icon: 'information-circle-outline' },
  warning: { background: colors.warningSoft, text: 'warning', icon: 'warning-outline' },
  danger: { background: colors.dangerSoft, text: 'danger', icon: 'alert-circle-outline' },
  success: { background: colors.successSoft, text: 'success', icon: 'checkmark-circle-outline' },
} as const;

export function Notice({ message, tone = 'info' }: { message: string; tone?: keyof typeof tones }) {
  const palette = tones[tone];
  return (
    <View
      style={[styles.notice, { backgroundColor: palette.background }]}
      accessibilityRole={tone === 'danger' ? 'alert' : 'text'}
      accessibilityLiveRegion="polite"
    >
      <Ionicons name={palette.icon} size={18} color={colors[palette.text]} />
      <AppText variant="caption" color={palette.text} style={styles.text}>
        {message}
      </AppText>
    </View>
  );
}

const styles = StyleSheet.create({
  notice: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm, borderRadius: radius.md, padding: spacing.md },
  text: { flex: 1 },
});
