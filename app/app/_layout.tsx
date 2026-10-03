import { Stack } from 'expo-router';
import { StatusBar } from 'expo-status-bar';

import { SessionProvider } from '@/auth/SessionProvider';
import { colors } from '@/theme';
import { installWebStyles } from '@/theme/webStyles';

installWebStyles();

export default function RootLayout() {
  return (
    <SessionProvider>
      <StatusBar style="light" />
      <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />
    </SessionProvider>
  );
}
