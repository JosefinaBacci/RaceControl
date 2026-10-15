import { Stack } from 'expo-router';

import { colors } from '@/theme';

// The list stays underneath the detail even when it is opened directly, so going
// back always returns to it.
export const unstable_settings = { anchor: 'index' };

export default function PublicDriversLayout() {
  return <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />;
}
