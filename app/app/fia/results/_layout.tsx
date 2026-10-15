import { Stack } from 'expo-router';

import { colors } from '@/theme';

// The race list stays underneath the editor even when it is opened directly, so
// going back always returns to it.
export const unstable_settings = { anchor: 'index' };

export default function ResultsLayout() {
  return <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: colors.background } }} />;
}
