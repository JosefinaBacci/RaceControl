import { Button, Screen } from '@/components';
import { useSanctions } from '@/data/hooks';
import { AcknowledgementBadge } from '@/features/AcknowledgementBadge';
import { MockNotice } from '@/features/MockNotice';
import { SanctionList } from '@/features/SanctionList';

export default function FiaResultsScreen() {
  const sanctions = useSanctions('all');

  return (
    <Screen
      title="Puntajes y sanciones"
      subtitle="Publicación de resultados y seguimiento del acuse de cada escudería."
      headerAction={<Button label="Publicar puntajes" icon="cloud-upload-outline" compact disabled />}
    >
      <MockNotice />
      <SanctionList title="Sanciones emitidas" sanctions={sanctions}
        trailing={(sanction) => <AcknowledgementBadge acknowledged={sanction.acknowledged} acknowledgedLabel="Acusada" />}
      />
    </Screen>
  );
}
