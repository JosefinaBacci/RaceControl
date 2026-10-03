import { Tabs } from 'expo-router';

import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { RoleGate } from '@/navigation/RoleGate';
import { headerLogo, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function FiaLayout() {
  const screenOptions = useTabScreenOptions();

  return (
    <RoleGate roles={['fia_admin']}>
      <Tabs screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}>
        <Tabs.Screen name="index" options={{ title: 'Panel', tabBarIcon: tabIcon('speedometer-outline') }} />
        <Tabs.Screen name="users" options={{ title: 'Usuarios', tabBarIcon: tabIcon('people-outline') }} />
        <Tabs.Screen name="calendar" options={{ title: 'Calendario', tabBarIcon: tabIcon('calendar-outline') }} />
        <Tabs.Screen name="results" options={{ title: 'Resultados', tabBarIcon: tabIcon('trophy-outline') }} />
        <Tabs.Screen name="notifications" options={{ title: 'Comunicados', tabBarIcon: tabIcon('megaphone-outline') }} />
      </Tabs>
    </RoleGate>
  );
}
