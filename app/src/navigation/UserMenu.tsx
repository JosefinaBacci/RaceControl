import Ionicons from '@expo/vector-icons/Ionicons';
import { router, type Href } from 'expo-router';
import { useState, type ComponentProps } from 'react';
import { Modal, Pressable, StyleSheet, View } from 'react-native';

import type { SessionUser } from '@/api/auth';
import { roleHome, roleLabel } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { AppText, Avatar } from '@/components';
import { RoleIcon } from '@/features/accounts/RoleIcon';
import { teamName } from '@/mocks/catalog';
import { colors, radius, spacing } from '@/theme';

type MenuItem = {
  label: string;
  icon: ComponentProps<typeof Ionicons>['name'];
  onSelect: () => void;
  isDestructive?: boolean;
};

const menuWidth = 260;

export function UserMenu({ user }: { user: SessionUser }) {
  const { signOut } = useSession();
  const [isOpen, setIsOpen] = useState(false);
  const subtitle = user.teamId === null ? roleLabel[user.role] : `${roleLabel[user.role]} · ${teamName(user.teamId)}`;

  const navigate = (href: Href) => {
    setIsOpen(false);
    router.push(href);
  };
  const leave = async () => {
    setIsOpen(false);
    try {
      await signOut();
    } finally {
      router.replace('/');
    }
  };

  const items: MenuItem[] = [
    { label: 'Mi panel', icon: 'grid-outline', onSelect: () => navigate(roleHome[user.role]) },
    { label: 'Mi cuenta', icon: 'person-circle-outline', onSelect: () => navigate('/account') },
    { label: 'Cerrar sesión', icon: 'log-out-outline', onSelect: leave, isDestructive: true },
  ];

  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`Menú de ${user.username}`}
        accessibilityState={{ expanded: isOpen }}
        onPress={() => setIsOpen(true)}
        hitSlop={8}
        style={({ pressed }) => [styles.trigger, pressed && styles.pressed]}
      >
        <Avatar name={user.username} size={32} color={colors.accent} />
        <Ionicons name="chevron-down" size={14} color={colors.textMuted} />
      </Pressable>
      <Modal visible={isOpen} transparent animationType="fade" onRequestClose={() => setIsOpen(false)}>
        <Pressable accessibilityLabel="Cerrar menú" style={styles.backdrop} onPress={() => setIsOpen(false)}>
          <View style={styles.menu} accessibilityRole="menu">
            <View style={styles.header}>
              <AppText variant="bodyStrong" numberOfLines={1}>
                {user.username}
              </AppText>
              <View style={styles.role}>
                <RoleIcon role={user.role} teamId={user.teamId} />
                <AppText variant="caption" color="textMuted" numberOfLines={1} style={styles.roleText}>
                  {subtitle}
                </AppText>
              </View>
            </View>
            {items.map((item) => (
              <Pressable
                key={item.label}
                accessibilityRole="menuitem"
                onPress={item.onSelect}
                style={({ pressed }) => [styles.item, pressed && styles.itemPressed]}
              >
                <Ionicons name={item.icon} size={18} color={item.isDestructive ? colors.danger : colors.text} />
                <AppText color={item.isDestructive ? 'danger' : 'text'}>{item.label}</AppText>
              </Pressable>
            ))}
          </View>
        </Pressable>
      </Modal>
    </>
  );
}

const styles = StyleSheet.create({
  trigger: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs, marginRight: spacing.lg },
  pressed: { opacity: 0.75 },
  backdrop: { flex: 1 },
  menu: {
    position: 'absolute',
    top: 56,
    right: spacing.lg,
    width: menuWidth,
    backgroundColor: colors.surfaceRaised,
    borderRadius: radius.md,
    borderWidth: 1,
    borderColor: colors.border,
    paddingVertical: spacing.xs,
  },
  header: { paddingHorizontal: spacing.lg, paddingVertical: spacing.md, gap: 2, borderBottomWidth: 1, borderBottomColor: colors.border },
  role: { flexDirection: 'row', alignItems: 'center', gap: spacing.xs },
  roleText: { flexShrink: 1 },
  item: { flexDirection: 'row', alignItems: 'center', gap: spacing.md, paddingHorizontal: spacing.lg, paddingVertical: spacing.md },
  itemPressed: { backgroundColor: colors.surfaceMuted },
});
