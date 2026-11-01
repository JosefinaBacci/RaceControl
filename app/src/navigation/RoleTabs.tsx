import { Tabs } from 'expo-router';

import { HeaderSessionButton } from './HeaderSessionButton';
import type { TabBadges, TabSpec } from './tabSpecs';
import { headerLogo, tabBadge, tabIcon, useTabScreenOptions } from './tabOptions';

type RoleTabsProps = {
  tabs: readonly TabSpec[];
  badges?: TabBadges;
  onTabFocus?: () => void;
};

export function RoleTabs({ tabs, badges = {}, onTabFocus }: RoleTabsProps) {
  const screenOptions = useTabScreenOptions();

  return (
    <Tabs
      screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}
      screenListeners={onTabFocus ? { focus: onTabFocus } : undefined}
    >
      {tabs.map((tab) => (
        <Tabs.Screen
          key={tab.name}
          name={tab.name}
          options={{ title: tab.title, tabBarIcon: tabIcon(tab.icon), tabBarBadge: tabBadge(badges[tab.name] ?? 0) }}
        />
      ))}
    </Tabs>
  );
}
