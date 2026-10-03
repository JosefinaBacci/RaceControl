import { Avatar, Button, Card, ListItem, Screen } from '@/components';
import { useSanctions, useScoreNotices } from '@/data/hooks';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { MockNotice } from '@/features/MockNotice';
import { SanctionList } from '@/features/SanctionList';
import { teamById } from '@/mocks/catalog';

export default function FiaResultsScreen() {
  const scoreNotices = useScoreNotices();
  const sanctions = useSanctions('all');

  return (
    <Screen
      title="Puntajes y sanciones"
      subtitle="Publicación de resultados y seguimiento del acuse de cada escudería."
      headerAction={<Button label="Publicar puntajes" icon="cloud-upload-outline" compact disabled />}
    >
      <MockNotice />
      <Card title="Puntajes publicados — GP de Singapur">
        {scoreNotices.map((notice, index) => {
          const team = teamById(notice.teamId);
          return (
            <ListItem
              key={notice.id}
              leading={<Avatar name={team.name} color={team.color} />}
              title={team.name}
              subtitle={`${notice.points} puntos`}
              trailing={<AcknowledgementBadge acknowledged={notice.acknowledged} />}
              isLast={index === scoreNotices.length - 1}
            />
          );
        })}
      </Card>
      <SanctionList title="Sanciones emitidas" sanctions={sanctions}
        trailing={(sanction) => <AcknowledgementBadge acknowledged={sanction.acknowledged} acknowledgedLabel="Acusada" />}
      />
    </Screen>
  );
}
