import Ionicons from '@expo/vector-icons/Ionicons';
import { router } from 'expo-router';
import { Pressable, StyleSheet } from 'react-native';

import { useSession } from '@/auth/SessionProvider';
import { AppText } from '@/components';
import { colors, radius, spacing } from '@/theme';

import { UserMenu } from './UserMenu';

export function HeaderSessionButton() {
  const session = useSession();

  if (session.status === 'authenticated') {
    return <UserMenu user={session.user} />;
  }
  return (
    <Pressable
      accessibilityRole="button"
      onPress={() => router.push('/login')}
      hitSlop={8}
      style={({ pressed }) => [styles.action, pressed && styles.pressed]}
    >
      <Ionicons name="person-circle-outline" size={18} color={colors.text} />
      <AppText variant="caption" style={styles.label}>
        Ingresar
      </AppText>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  action: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    marginRight: spacing.lg,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: radius.pill,
    backgroundColor: colors.accent,
  },
  pressed: { opacity: 0.75 },
  label: { fontWeight: '700' },
});
