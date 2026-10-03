import { useSession } from '@/auth/SessionProvider';
import { Badge, Button, Card, ListItem, Screen } from '@/components';
import { useTeamInbox } from '@/data/hooks';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { MockNotice } from '@/features/MockNotice';

export default function TeamInboxScreen() {
  const { items, acknowledge } = useTeamInbox(useSession().user?.teamId ?? null);

  return (
    <Screen title="Acuses de recibo" subtitle="Confirmá a la FIA que tomaste conocimiento de sanciones y puntajes.">
      <MockNotice />
      <Card title="Sanciones y puntajes">
        {items.map((item, index) => (
          <ListItem
            key={item.key}
            leading={<Badge label={item.kind} tone={item.kind === 'Sanción' ? 'danger' : 'info'} />}
            title={item.title}
            subtitle={item.detail}
            trailing={
              item.acknowledged ? (
                <AcknowledgementBadge acknowledged />
              ) : (
                <Button label="Acusar recibo" icon="checkmark-done" compact onPress={() => acknowledge(item.key)} />
              )
            }
            isLast={index === items.length - 1}
          />
        ))}
      </Card>
    </Screen>
  );
}
