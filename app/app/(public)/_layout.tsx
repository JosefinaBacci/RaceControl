import { Tabs } from 'expo-router';

import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { headerLogo, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function PublicLayout() {
  const screenOptions = useTabScreenOptions();

  return (
    <Tabs screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}>
      <Tabs.Screen name="index" options={{ title: 'Inicio', tabBarIcon: tabIcon('home-outline') }} />
      <Tabs.Screen name="calendar" options={{ title: 'Carreras', tabBarIcon: tabIcon('flag-outline') }} />
      <Tabs.Screen name="standings" options={{ title: 'Puntajes', tabBarIcon: tabIcon('trophy-outline') }} />
      <Tabs.Screen name="sanctions" options={{ title: 'Sanciones', tabBarIcon: tabIcon('alert-circle-outline') }} />
    </Tabs>
  );
}
