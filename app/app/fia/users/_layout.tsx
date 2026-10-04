import { Stack } from 'expo-router';

import { colors } from '@/theme';

// The list stays underneath the create and edit screens even when they are opened directly or
// from another tab, so going back always returns to it.
export const unstable_settings = { anchor: 'index' };

export default function UsersLayout() {
  return <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />;
}
