import { Tabs } from 'expo-router';

import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { RoleGate } from '@/navigation/RoleGate';
import { headerLogo, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function TeamLayout() {
  const screenOptions = useTabScreenOptions();

  return (
    <RoleGate role="team_admin">
      <Tabs screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton mode="private" /> }}>
        <Tabs.Screen name="index" options={{ title: 'Mi escudería', tabBarIcon: tabIcon('speedometer-outline') }} />
        <Tabs.Screen name="drivers" options={{ title: 'Pilotos', tabBarIcon: tabIcon('person-outline') }} />
        <Tabs.Screen name="inbox" options={{ title: 'Acuses', tabBarIcon: tabIcon('checkmark-done-outline') }} />
        <Tabs.Screen name="notifications" options={{ title: 'Comunicados', tabBarIcon: tabIcon('mail-outline') }} />
      </Tabs>
    </RoleGate>
  );
}
