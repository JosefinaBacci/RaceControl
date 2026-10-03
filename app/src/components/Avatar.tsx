import { Image, StyleSheet, View, type ImageSourcePropType } from 'react-native';

import { colors } from '@/theme';

import { AppText } from './AppText';

function initialsOf(name: string): string {
  return name
    .split(/[\s._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('');
}

type AvatarProps = {
  name: string;
  color?: string;
  size?: number;
  image?: ImageSourcePropType;
};

export function Avatar({ name, color = colors.surfaceMuted, size = 36, image }: AvatarProps) {
  const shape = { width: size, height: size, borderRadius: size / 2 };
  if (image) {
    return <Image source={image} accessibilityLabel={name} style={[styles.image, shape]} />;
  }
  return (
    <View style={[styles.avatar, shape, { borderColor: color }]}>
      <AppText variant="caption" style={{ fontSize: size * 0.36, fontWeight: '800' }}>
        {initialsOf(name)}
      </AppText>
    </View>
  );
}

const styles = StyleSheet.create({
  avatar: {
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.surfaceRaised,
    borderWidth: 2,
  },
  image: { backgroundColor: colors.surfaceRaised },
});
