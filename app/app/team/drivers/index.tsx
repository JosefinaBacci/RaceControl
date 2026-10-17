import { router } from 'expo-router';

import { useSession } from '@/auth/SessionProvider';
import { AppText, Button, Screen } from '@/components';
import { useTeamDrivers } from '@/data/drivers';
import { DriverCard, DriverGrid } from '@/features/DriverCard';
import { RemoteContent } from '@/features/RemoteContent';
import { teamById } from '@/mocks/catalog';

export default function TeamDriversScreen() {
  const teamId = useSession().user?.teamId ?? null;
  const team = teamById(teamId);
  const { state, reload } = useTeamDrivers();

  return (
    <Screen
      title="Pilotos"
      subtitle={`Titulares y suplentes de ${team.name}.`}
      headerAction={<Button label="Agregar piloto" icon="add" compact onPress={() => router.push('/team/drivers/new')} />}
    >
      <Button label="Buscar pilotos de todas las escuderías" icon="search" variant="secondary"
        onPress={() => router.push('/drivers')} />
      <RemoteContent state={state} onRetry={reload}>
        {(drivers) =>
          drivers.length === 0 ? (
            <AppText color="textMuted">Todavía no cargaste pilotos.</AppText>
          ) : (
            <DriverGrid>
              {drivers.map((driver) => (
                <DriverCard
                  key={driver.id}
                  driver={driver}
                  teamColor={team.color}
                  onPress={() => router.push({ pathname: '/team/drivers/[id]', params: { id: String(driver.id) } })}
                />
              ))}
            </DriverGrid>
          )
        }
      </RemoteContent>
    </Screen>
  );
}
