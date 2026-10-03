import { Button, Screen } from '@/components';
import { useNotifications } from '@/data/hooks';
import { MockNotice } from '@/features/MockNotice';
import { NotificationList } from '@/features/NotificationList';

export default function FiaNotificationsScreen() {
  const notifications = useNotifications();

  return (
    <Screen
      title="Comunicados"
      subtitle="Mensajes de la FIA a las escuderías y su confirmación de lectura."
      headerAction={<Button label="Nuevo comunicado" icon="add" compact disabled />}
    >
      <MockNotice />
      <NotificationList notifications={notifications} audience="sender" />
    </Screen>
  );
}
