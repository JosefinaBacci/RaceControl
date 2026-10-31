import { Stack } from 'expo-router';

import { colors } from '@/theme';

// The list stays underneath the register and detail screens even when they are opened directly or
// from another tab, so going back always returns to it.
export const unstable_settings = { anchor: 'index' };

export default function FiaSanctionsLayout() {
  return <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />;
}
