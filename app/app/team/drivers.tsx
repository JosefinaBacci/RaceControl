import { useSession } from '@/auth/SessionProvider';
import { Button, Screen } from '@/components';
import { useTeamDrivers } from '@/data/hooks';
import { DriverCard, DriverGrid } from '@/features/DriverCard';
import { MockNotice } from '@/features/MockNotice';
import { teamById } from '@/mocks/catalog';

export default function TeamDriversScreen() {
  const teamId = useSession().user?.teamId ?? null;
  const team = teamById(teamId);
  const drivers = useTeamDrivers(teamId);

  return (
    <Screen
      title="Pilotos"
      subtitle={`Titulares y suplentes de ${team.name}.`}
      headerAction={<Button label="Agregar piloto" icon="add" compact disabled />}
    >
      <MockNotice />
      <DriverGrid>
        {drivers.map((driver) => (
          <DriverCard key={driver.id} driver={driver} teamColor={team.color} />
        ))}
      </DriverGrid>
    </Screen>
  );
}
