import Ionicons from '@expo/vector-icons/Ionicons';
import { Image, StyleSheet } from 'react-native';

import type { Role } from '@/api/auth';
import { accountLogo } from '@/logos';
import { colors } from '@/theme';

const iconSize = 18;

const fallbackIcons = {
  fia_admin: 'shield-checkmark-outline',
  team_admin: 'car-sport-outline',
} as const satisfies Record<Role, string>;

export function RoleIcon({ role, teamId }: { role: Role; teamId: number | null }) {
  const logo = accountLogo(role, teamId);
  if (logo) {
    return <Image source={logo} style={styles.logo} accessibilityIgnoresInvertColors />;
  }
  return <Ionicons name={fallbackIcons[role]} size={iconSize - 2} color={colors.textMuted} />;
}

const styles = StyleSheet.create({
  logo: { width: iconSize, height: iconSize, borderRadius: iconSize / 2 },
});
