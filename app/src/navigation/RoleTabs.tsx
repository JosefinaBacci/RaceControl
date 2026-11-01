import { Tabs, useSegments } from 'expo-router';

import { colors } from '@/theme';

import { HeaderSessionButton } from './HeaderSessionButton';
import { moreMenuBadge, moreMenuTabs, moreTab, phoneTabLabel, type TabBadges, type TabSpec } from './tabSpecs';
import { headerLogo, tabBadge, tabIcon, useShowsSidebar, useTabScreenOptions } from './tabOptions';

type RoleTabsProps = {
  tabs: readonly TabSpec[];
  badges?: TabBadges;
  onTabFocus?: () => void;
};

// The desktop sidebar lists every tab; a phone shows the primary ones plus "Más",
// which lists the rest (see the more.tsx route of each group).
export function RoleTabs({ tabs, badges = {}, onTabFocus }: RoleTabsProps) {
  const screenOptions = useTabScreenOptions();
  const showsSidebar = useShowsSidebar();
  const isMenuTabFocused = useIsMenuTabFocused(tabs);

  return (
    <Tabs
      screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}
      screenListeners={onTabFocus ? { focus: onTabFocus } : undefined}
    >
      {tabs.map((tab) => (
        <Tabs.Screen
          key={tab.name}
          name={tab.name}
          options={{
            title: tab.title,
            tabBarLabel: showsSidebar ? tab.title : phoneTabLabel(tab),
            tabBarIcon: tabIcon(tab.icon),
            tabBarBadge: tabBadge(badges[tab.name] ?? 0),
            href: showsSidebar || tab.isPhonePrimary ? undefined : null,
          }}
        />
      ))}
      <Tabs.Screen
        name={moreTab.name}
        options={{
          title: moreTab.title,
          tabBarIcon: tabIcon(moreTab.icon),
          tabBarBadge: tabBadge(moreMenuBadge(tabs, badges)),
          tabBarInactiveTintColor: isMenuTabFocused ? colors.accent : colors.textMuted,
          href: showsSidebar ? null : undefined,
        }}
      />
    </Tabs>
  );
}

// A screen reached from "Más" has no tab of its own on a phone, so "Más" stays lit.
function useIsMenuTabFocused(tabs: readonly TabSpec[]): boolean {
  const segments: string[] = useSegments();
  const focusedTabName = segments[1] ?? 'index';
  return moreMenuTabs(tabs).some((tab) => tab.name === focusedTabName);
}
