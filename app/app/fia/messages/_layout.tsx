import { Stack } from 'expo-router';

import { colors } from '@/theme';

// The conversation list stays underneath a conversation even when it is opened
// directly, so going back always returns to it.
export const unstable_settings = { anchor: 'index' };

export default function MessagesLayout() {
  return <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />;
}
