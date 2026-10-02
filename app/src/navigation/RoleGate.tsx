import { Redirect } from 'expo-router';
import type { ReactNode } from 'react';
import { ActivityIndicator, StyleSheet, View } from 'react-native';

import type { Role } from '@/api/auth';
import { roleHome } from '@/auth/roles';
import { useSession } from '@/auth/SessionProvider';
import { colors } from '@/theme';

export function RoleGate({ role, children }: { role: Role; children: ReactNode }) {
  const session = useSession();

  if (session.status === 'loading') {
    return <FullScreenSpinner />;
  }
  if (session.status === 'anonymous') {
    return <Redirect href="/login" />;
  }
  if (session.user.role !== role) {
    return <Redirect href={roleHome[session.user.role]} />;
  }
  return children;
}

export function FullScreenSpinner() {
  return (
    <View style={styles.center}>
      <ActivityIndicator size="large" color={colors.accent} />
    </View>
  );
}

const styles = StyleSheet.create({
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.background },
});
