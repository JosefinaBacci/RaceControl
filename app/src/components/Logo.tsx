import { StyleSheet, View } from 'react-native';

import { colors } from '@/theme';

import { AppText } from './AppText';

const sizes = {
  sm: { bar: 14, gap: 3, word: 16, tagline: 0 },
  lg: { bar: 34, gap: 6, word: 34, tagline: 11 },
} as const;

export function Logo({ size = 'sm', showTagline = false }: { size?: keyof typeof sizes; showTagline?: boolean }) {
  const dimensions = sizes[size];

  return (
    <View style={styles.container} accessibilityRole="image" accessibilityLabel="RaceControl">
      <View style={styles.row}>
        <View style={[styles.bars, { gap: dimensions.gap }]}>
          {[1, 0.75, 0.5].map((widthFactor) => (
            <View
              key={widthFactor}
              style={[styles.bar, { height: dimensions.bar / 3.4, width: dimensions.bar * 1.6 * widthFactor }]}
            />
          ))}
        </View>
        <AppText style={[styles.word, { fontSize: dimensions.word }]}>
          RACE<AppText style={[styles.word, styles.wordAccent, { fontSize: dimensions.word }]}>CONTROL</AppText>
        </AppText>
      </View>
      {showTagline && dimensions.tagline > 0 ? (
        <AppText style={[styles.tagline, { fontSize: dimensions.tagline }]}>SPEED · DATA · PASSION</AppText>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: { alignItems: 'flex-start', gap: 10 },
  row: { flexDirection: 'row', alignItems: 'center', gap: 10 },
  bars: { transform: [{ skewX: '-24deg' }] },
  bar: { backgroundColor: colors.accent, borderRadius: 2 },
  word: { fontWeight: '900', fontStyle: 'italic', letterSpacing: -0.5, color: colors.text },
  wordAccent: { color: colors.accent },
  tagline: { color: colors.textMuted, fontWeight: '700', letterSpacing: 6 },
});
