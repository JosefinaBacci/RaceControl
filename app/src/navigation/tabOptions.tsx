import Ionicons from '@expo/vector-icons/Ionicons';
import { Platform, useWindowDimensions, type ColorValue } from 'react-native';

import { Logo } from '@/components';
import { colors } from '@/theme';

import type { TabIconName } from './tabSpecs';

const sidebarBreakpoint = 1024;

const baseTabOptions = {
  headerStyle: { backgroundColor: colors.background, borderBottomColor: colors.border, borderBottomWidth: 1 },
  headerShadowVisible: false,
  headerTintColor: colors.text,
  headerTitleAlign: 'left' as const,
  headerTitleStyle: { fontWeight: '800' as const, color: colors.text },
  tabBarActiveTintColor: colors.accent,
  tabBarInactiveTintColor: colors.textMuted,
  tabBarActiveBackgroundColor: 'transparent',
  sceneStyle: { backgroundColor: colors.background },
};

export function useShowsSidebar(): boolean {
  const { width } = useWindowDimensions();
  return Platform.OS === 'web' && width >= sidebarBreakpoint;
}

export function useTabScreenOptions() {
  if (useShowsSidebar()) {
    return {
      ...baseTabOptions,
      tabBarPosition: 'left' as const,
      tabBarVariant: 'material' as const,
      tabBarLabelPosition: 'beside-icon' as const,
      tabBarActiveBackgroundColor: colors.accent,
      tabBarActiveTintColor: colors.textInverse,
      tabBarStyle: { backgroundColor: colors.surface, borderRightColor: colors.border, width: 240, paddingTop: 12 },
      tabBarItemStyle: { borderRadius: 10, marginHorizontal: 12, marginVertical: 2 },
    };
  }
  return {
    ...baseTabOptions,
    tabBarStyle: { backgroundColor: colors.surface, borderTopColor: colors.border },
  };
}

export function headerLogo() {
  return <Logo size="sm" />;
}

export function tabIcon(name: TabIconName) {
  return function TabIcon({ color, size }: { color: ColorValue; size: number }) {
    return <Ionicons name={name} color={color} size={size} />;
  };
}

// An empty count hides the badge instead of showing a zero.
export function tabBadge(count: number): number | undefined {
  return count > 0 ? count : undefined;
}
