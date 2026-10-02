import { useEffect, useState } from 'react';
import { StyleSheet, View } from 'react-native';

import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

const second = 1000;
const units = [
  { label: 'Días', milliseconds: 86_400 * second },
  { label: 'Horas', milliseconds: 3_600 * second },
  { label: 'Min', milliseconds: 60 * second },
  { label: 'Seg', milliseconds: second },
] as const;

function splitRemaining(remaining: number) {
  let rest = Math.max(remaining, 0);
  return units.map((unit) => {
    const value = Math.floor(rest / unit.milliseconds);
    rest -= value * unit.milliseconds;
    return { label: unit.label, value };
  });
}

export function Countdown({ target }: { target: Date }) {
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const timer = setInterval(() => setNow(Date.now()), second);
    return () => clearInterval(timer);
  }, []);

  return (
    <View style={styles.box} accessibilityRole="timer">
      <AppText variant="caption" color="textMuted">
        Comienza en
      </AppText>
      <View style={styles.row}>
        {splitRemaining(target.getTime() - now).map((part) => (
          <View key={part.label} style={styles.unit}>
            <AppText variant="display">{String(part.value).padStart(2, '0')}</AppText>
            <AppText variant="overline" color="textMuted">
              {part.label}
            </AppText>
          </View>
        ))}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  box: {
    backgroundColor: colors.surface,
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.border,
    padding: spacing.lg,
    gap: spacing.sm,
  },
  row: { flexDirection: 'row', justifyContent: 'space-around' },
  unit: { alignItems: 'center', minWidth: 56 },
});
