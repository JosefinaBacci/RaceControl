import { Tabs } from 'expo-router';

import { TeamScoreNoticesProvider, usePendingScoreNotices, useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { UnreadMessagesProvider, useUnreadMessages } from '@/data/UnreadMessagesProvider';
import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { RoleGate } from '@/navigation/RoleGate';
import { headerLogo, tabBadge, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function TeamLayout() {
  return (
    <RoleGate roles={['team_admin']}>
      <TeamScoreNoticesProvider>
        <UnreadMessagesProvider>
          <TeamTabs />
        </UnreadMessagesProvider>
      </TeamScoreNoticesProvider>
    </RoleGate>
  );
}

function TeamTabs() {
  const screenOptions = useTabScreenOptions();
  const pendingScores = usePendingScoreNotices();
  const { reload: reloadNotices } = useTeamScoreNotices();
  const { unreadCount, reload: reloadUnread } = useUnreadMessages();

  // Reloading on every tab focus keeps the badges current wherever the team admin
  // is, without polling: a change the FIA publishes or a message it sends shows
  // up on the next tab.
  const reloadBadges = () => {
    void reloadNotices();
    void reloadUnread();
  };

  return (
    <Tabs
      screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}
      screenListeners={{ focus: reloadBadges }}
    >
      <Tabs.Screen name="index" options={{ title: 'Mi escudería', tabBarIcon: tabIcon('speedometer-outline') }} />
      <Tabs.Screen name="drivers" options={{ title: 'Pilotos', tabBarIcon: tabIcon('person-outline') }} />
      <Tabs.Screen name="scores" options={{ title: 'Puntajes', tabBarIcon: tabIcon('trophy-outline') }} />
      <Tabs.Screen name="inbox" options={{ title: 'Acuses', tabBarIcon: tabIcon('checkmark-done-outline') }} />
      <Tabs.Screen
        name="notifications"
        options={{ title: 'Notificaciones', tabBarIcon: tabIcon('notifications-outline'), tabBarBadge: tabBadge(pendingScores) }}
      />
      <Tabs.Screen
        name="messages"
        options={{ title: 'Mensajes', tabBarIcon: tabIcon('chatbubbles-outline'), tabBarBadge: tabBadge(unreadCount) }}
      />
      <Tabs.Screen name="public" options={{ title: 'Sitio público', tabBarIcon: tabIcon('globe-outline') }} />
      <Tabs.Screen name="assignments" options={{ title: 'Asignaciones', tabBarIcon: tabIcon('time-outline') }} />
    </Tabs>
  );
}
