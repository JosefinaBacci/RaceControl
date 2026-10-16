import { Tabs } from 'expo-router';

import { TeamScoreNoticesProvider, usePendingScoreNotices, useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
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
  const { reload: reloadNotices } = useTeamScoreNotices();

  // Reloading on every tab focus keeps the badge current wherever the team admin
  // is, without polling: a change the FIA publishes shows up on the next tab.
  return (
    <Tabs
      screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}
      screenListeners={{ focus: () => void reloadNotices() }}
    >
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
