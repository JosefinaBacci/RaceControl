import Ionicons from '@expo/vector-icons/Ionicons';
import { StyleSheet, View } from 'react-native';

import { Badge, Card, ListItem } from '@/components';
import { formatEventDate } from '@/format';
import type { Notification } from '@/mocks/data';
import { colors, radius } from '@/theme';

type NotificationListProps = {
  notifications: Notification[];
  audience: 'sender' | 'recipient';
};

export function NotificationList({ notifications, audience }: NotificationListProps) {
  return (
    <Card title={audience === 'sender' ? 'Enviados' : 'Recibidos'}>
      {notifications.map((notification, index) => (
        <ListItem
          key={notification.id}
          leading={
            <View style={styles.icon}>
              <Ionicons name={audience === 'sender' ? 'megaphone-outline' : 'mail-unread-outline'} size={18} color={colors.accent} />
            </View>
          }
          title={notification.title}
          subtitle={`${formatEventDate(notification.sentAt)} · ${notification.body}`}
          trailing={
            audience === 'sender' ? (
              <Badge label={`${notification.acknowledgedBy}/${notification.recipients} leídos`} tone="info" icon="eye-outline" />
            ) : null
          }
          isLast={index === notifications.length - 1}
        />
      ))}
    </Card>
  );
}

const styles = StyleSheet.create({
  icon: {
    width: 38,
    height: 38,
    borderRadius: radius.md,
    backgroundColor: colors.accentSoft,
    alignItems: 'center',
    justifyContent: 'center',
  },
});
