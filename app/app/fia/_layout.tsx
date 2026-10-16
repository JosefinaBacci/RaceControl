import { Tabs } from 'expo-router';

import { UnreadMessagesProvider, useUnreadMessages } from '@/data/UnreadMessagesProvider';
import { HeaderSessionButton } from '@/navigation/HeaderSessionButton';
import { RoleGate } from '@/navigation/RoleGate';
import { headerLogo, tabBadge, tabIcon, useTabScreenOptions } from '@/navigation/tabOptions';

export default function FiaLayout() {
  return (
    <RoleGate roles={['fia_admin']}>
      <UnreadMessagesProvider>
        <FiaTabs />
      </UnreadMessagesProvider>
    </RoleGate>
  );
}

function FiaTabs() {
  const screenOptions = useTabScreenOptions();
  const { unreadCount, reload: reloadUnread } = useUnreadMessages();

  // Reloading on every tab focus keeps the badge current without polling: a
  // message a team sends shows up on the next tab.
  return (
    <Tabs
      screenOptions={{ ...screenOptions, headerTitle: headerLogo, headerRight: () => <HeaderSessionButton /> }}
      screenListeners={{ focus: () => void reloadUnread() }}
    >
      <Tabs.Screen name="index" options={{ title: 'Panel', tabBarIcon: tabIcon('speedometer-outline') }} />
      <Tabs.Screen name="users" options={{ title: 'Usuarios', tabBarIcon: tabIcon('people-outline') }} />
      <Tabs.Screen name="calendar" options={{ title: 'Calendario', tabBarIcon: tabIcon('calendar-outline') }} />
      <Tabs.Screen name="results" options={{ title: 'Resultados', tabBarIcon: tabIcon('trophy-outline') }} />
      <Tabs.Screen name="notifications" options={{ title: 'Acuses', tabBarIcon: tabIcon('checkmark-done-outline') }} />
      <Tabs.Screen
        name="messages"
        options={{ title: 'Mensajes', tabBarIcon: tabIcon('chatbubbles-outline'), tabBarBadge: tabBadge(unreadCount) }}
      />
    </Tabs>
  );
}
