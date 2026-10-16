import { Tabs } from 'expo-router';

import { TeamScoreNoticesProvider, usePendingScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { RoleGate } from '@/navigation/RoleGate';
import { headerLogo, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function TeamLayout() {
  return (
    <RoleGate roles={['team_admin']}>
      <TeamScoreNoticesProvider>
        <TeamTabs />
      </TeamScoreNoticesProvider>
    </RoleGate>
  );
}

function TeamTabs() {
  const screenOptions = useTabScreenOptions();
  const pendingScores = usePendingScoreNotices();

  return (
    <Tabs screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}>
      <Tabs.Screen name="index" options={{ title: 'Mi escudería', tabBarIcon: tabIcon('speedometer-outline') }} />
      <Tabs.Screen name="drivers" options={{ title: 'Pilotos', tabBarIcon: tabIcon('person-outline') }} />
      <Tabs.Screen name="scores" options={{ title: 'Puntajes', tabBarIcon: tabIcon('trophy-outline') }} />
      <Tabs.Screen name="inbox" options={{ title: 'Acuses', tabBarIcon: tabIcon('checkmark-done-outline') }} />
      <Tabs.Screen
        name="notifications"
        options={{
          title: 'Notificaciones',
          tabBarIcon: tabIcon('notifications-outline'),
          tabBarBadge: pendingScores > 0 ? pendingScores : undefined,
        }}
      />
    </Tabs>
  );
}
