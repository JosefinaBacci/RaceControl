import { LinearGradient } from 'expo-linear-gradient';
import type { ReactNode } from 'react';
import { Image, StyleSheet, View } from 'react-native';

import { Flag, type FlagCode } from '@/illustrations';
import type { Photo } from '@/photos';
import { colors, radius, spacing } from '@/theme';

import { AppText } from './AppText';

type HeroCardProps = {
  overline: string;
  title: string;
  subtitle?: string;
  accentColor?: string;
  flag?: FlagCode;
  illustration?: ReactNode;
  illustrationSide?: 'left' | 'right';
  photo?: Photo;
  children?: ReactNode;
};

export function HeroCard({ overline, title, subtitle, accentColor = colors.accent, flag, illustration, illustrationSide = 'right', photo, children }: HeroCardProps) {
  return (
    <View style={[styles.card, photo ? styles.cardTall : null, illustration ? styles.cardWithIllustration : null]}>
      {photo ? <PhotoBackdrop photo={photo} accentColor={accentColor} /> : <GradientBackdrop accentColor={accentColor} />}
      {illustration ? (
        <View style={[styles.illustration, illustrationSide === 'left' ? styles.illustrationLeft : styles.illustrationRight]} pointerEvents="none">
          {illustration}
        </View>
      ) : null}
      <View style={styles.content}>
        <View style={styles.overline}>
          {flag ? <Flag code={flag} width={22} /> : <View style={[styles.flag, { backgroundColor: accentColor }]} />}
          <AppText variant="overline" color="text">
            {overline}
          </AppText>
        </View>
        <AppText variant="title" accessibilityRole="header" style={styles.shadowed}>
          {title}
        </AppText>
        {subtitle ? <AppText style={[styles.subtitle, styles.shadowed]}>{subtitle}</AppText> : null}
        {children}
      </View>
      {photo ? (
        <AppText variant="caption" style={styles.credit}>
          {photo.credit}
        </AppText>
      ) : null}
    </View>
  );
}

function GradientBackdrop({ accentColor }: { accentColor: string }) {
  return (
    <>
      <LinearGradient
        colors={[accentColor, '#3A0A0A', colors.surface]}
        locations={[0, 0.45, 1]}
        start={{ x: 1, y: 0 }}
        end={{ x: 0, y: 1 }}
        style={StyleSheet.absoluteFill}
      />
      <View style={styles.stripes}>
        {[0, 1, 2, 3].map((index) => (
          <View key={index} style={[styles.stripe, { opacity: 0.08 + index * 0.04 }]} />
        ))}
      </View>
    </>
  );
}

function PhotoBackdrop({ photo, accentColor }: { photo: Photo; accentColor: string }) {
  return (
    <>
      <Image source={photo.source} style={styles.photo} resizeMode="cover" accessibilityIgnoresInvertColors />
      <LinearGradient
        colors={['rgba(11,11,15,0)', 'rgba(11,11,15,0.15)', 'rgba(11,11,15,0.92)']}
        locations={[0, 0.5, 1]}
        style={StyleSheet.absoluteFill}
      />
      <LinearGradient
        colors={[`${accentColor}40`, 'transparent']}
        start={{ x: 0, y: 1 }}
        end={{ x: 0.7, y: 0.3 }}
        style={StyleSheet.absoluteFill}
      />
    </>
  );
}

const styles = StyleSheet.create({
  photo: { position: 'absolute', top: 0, left: 0, width: '100%', height: '100%' },
  card: {
    minHeight: 220,
    borderRadius: radius.lg,
    overflow: 'hidden',
    borderWidth: 1,
    borderColor: colors.border,
    justifyContent: 'flex-end',
    backgroundColor: colors.surface,
  },
  cardTall: { minHeight: 310 },
  cardWithIllustration: { minHeight: 340 },
  stripes: {
    position: 'absolute',
    top: -40,
    right: -20,
    flexDirection: 'row',
    gap: 14,
    transform: [{ skewX: '-24deg' }],
  },
  stripe: { width: 26, height: 260, backgroundColor: colors.text },
  illustration: { position: 'absolute', top: spacing.lg, opacity: 0.95 },
  illustrationLeft: { left: spacing.xl },
  illustrationRight: { right: spacing.lg },
  content: { padding: spacing.xl, gap: spacing.sm },
  overline: { flexDirection: 'row', alignItems: 'center', gap: spacing.sm },
  flag: { width: 18, height: 12, borderRadius: 2 },
  subtitle: { color: '#D4D4DC' },
  shadowed: { textShadowColor: 'rgba(0,0,0,0.6)', textShadowRadius: 8, textShadowOffset: { width: 0, height: 1 } },
  credit: { position: 'absolute', bottom: spacing.sm, right: spacing.md, color: 'rgba(255,255,255,0.6)', fontSize: 10 },
});
