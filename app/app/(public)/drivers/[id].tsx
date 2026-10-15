import { router, useLocalSearchParams } from 'expo-router';

import { Badge, Button, Card, FieldGroup, Screen, StatRow, StatTile } from '@/components';
import { usePublicDriver } from '@/data/drivers';
import { RemoteContent } from '@/features/RemoteContent';
import { Flag } from '@/illustrations';

export default function DriverDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload } = usePublicDriver(Number(id));

  return (
    <Screen
      title="Piloto"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(driver) => (
          <Card
            title={`${driver.firstName} ${driver.lastName}`}
            action={<Badge label={driver.isStarter ? 'Titular' : 'Suplente'} tone={driver.isStarter ? 'accent' : 'neutral'} />}
          >
            <StatRow>
              <StatTile label="Número" value={driver.number ?? '—'} />
              <StatTile label="Escudería" value={driver.teamName} />
              <StatTile label="Categoría" value={driver.categoryName} />
            </StatRow>
            <FieldGroup label="País">
              <Flag code={driver.countryCode} width={36} />
            </FieldGroup>
          </Card>
        )}
      </RemoteContent>
    </Screen>
  );
}
