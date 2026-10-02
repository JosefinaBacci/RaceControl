import Ionicons from '@expo/vector-icons/Ionicons';
import { router } from 'expo-router';
import { useState, type ComponentProps } from 'react';
import { Pressable, StyleSheet } from 'react-native';

import { roleHome } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { AppText } from '@/components';
import { colors, radius, spacing } from '@/theme';

export function HeaderSessionButton({ mode }: { mode: 'public' | 'private' }) {
  const session = useSession();
  const [isSigningOut, setIsSigningOut] = useState(false);

  if (mode === 'private') {
    const signOut = async () => {
      setIsSigningOut(true);
      try {
        await session.signOut();
      } finally {
        setIsSigningOut(false);
        router.replace('/');
      }
    };
    return (
      <HeaderAction
        icon="log-out-outline"
        label={isSigningOut ? 'Saliendo…' : 'Cerrar sesión'}
        onPress={signOut}
        disabled={isSigningOut}
      />
    );
  }

  if (session.status === 'authenticated') {
    return <HeaderAction icon="grid-outline" label="Mi panel" onPress={() => router.push(roleHome[session.user.role])} highlighted />;
  }
  return <HeaderAction icon="person-circle-outline" label="Ingresar" onPress={() => router.push('/login')} highlighted />;
}

type HeaderActionProps = {
  icon: ComponentProps<typeof Ionicons>['name'];
  label: string;
  onPress: () => void;
  disabled?: boolean;
  highlighted?: boolean;
};

function HeaderAction({ icon, label, onPress, disabled, highlighted = false }: HeaderActionProps) {
  return (
    <Pressable
      accessibilityRole="button"
      onPress={onPress}
      disabled={disabled}
      hitSlop={8}
      style={({ pressed }) => [styles.action, highlighted && styles.highlighted, pressed && styles.pressed]}
    >
      <Ionicons name={icon} size={18} color={colors.text} />
      <AppText variant="caption" style={styles.label}>
        {label}
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
    borderWidth: 1,
    borderColor: colors.border,
  },
  highlighted: { backgroundColor: colors.accent, borderColor: colors.accent },
  pressed: { opacity: 0.75 },
  label: { fontWeight: '700' },
});
