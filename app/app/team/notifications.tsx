import { Screen } from '@/components';
import { useNotifications } from '@/data/hooks';
import { MockNotice } from '@/features/MockNotice';
import { NotificationList } from '@/features/NotificationList';

export default function TeamNotificationsScreen() {
  const notifications = useNotifications();

  return (
    <Screen title="Comunicados FIA" subtitle="Avisos y reglamentos enviados por la FIA.">
      <MockNotice />
      <NotificationList notifications={notifications} audience="recipient" />
    </Screen>
  );
}
