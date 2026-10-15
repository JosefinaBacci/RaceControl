import { router } from 'expo-router';
import { useState } from 'react';

import type { PublicDriver } from '@/api/drivers';
import { AppText, Card, ChipGroup, Screen, TextField } from '@/components';
import { usePublicDrivers } from '@/data/drivers';
import { DriverCard, DriverGrid } from '@/features/DriverCard';
import { categoryFilterOptions } from '@/features/filters';
import { RemoteContent } from '@/features/RemoteContent';
import { teamById, type CategoryFilter } from '@/mocks/catalog';

type TeamLineup = { teamId: number; teamName: string; drivers: PublicDriver[] };

export default function PublicDriversScreen() {
  const [search, setSearch] = useState('');
  const [category, setCategory] = useState<CategoryFilter>('all');
  const { state, reload } = usePublicDrivers({
    search,
    category: category === 'all' ? undefined : category,
  });

  return (
    <Screen title="Pilotos" subtitle="La alineación vigente de cada escudería.">
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <TextField
        label="Buscar piloto…"
        icon="search-outline"
        value={search}
        onChangeText={setSearch}
        autoCapitalize="none"
        autoCorrect={false}
      />
      <RemoteContent state={state} onRetry={reload}>
        {(drivers) =>
          drivers.length === 0 ? (
            <AppText color="textMuted">No hay pilotos que coincidan con la búsqueda.</AppText>
          ) : (
            groupByTeam(drivers).map((lineup) => (
              <Card key={lineup.teamId} title={lineup.teamName}>
                <DriverGrid>
                  {lineup.drivers.map((driver) => (
                    <DriverCard
                      key={driver.id}
                      driver={driver}
                      teamColor={teamById(lineup.teamId).color}
                      onPress={() => router.push({ pathname: '/drivers/[id]', params: { id: String(driver.id) } })}
                    />
                  ))}
                </DriverGrid>
              </Card>
            ))
          )
        }
      </RemoteContent>
    </Screen>
  );
}

// The API answers one flat list ordered by championship and team, so grouping is a
// single pass that keeps that order.
function groupByTeam(drivers: PublicDriver[]): TeamLineup[] {
  const lineups: TeamLineup[] = [];
  for (const driver of drivers) {
    const current = lineups[lineups.length - 1];
    if (current && current.teamId === driver.teamId) {
      current.drivers.push(driver);
    } else {
      lineups.push({ teamId: driver.teamId, teamName: driver.teamName, drivers: [driver] });
    }
  }
  return lineups;
}
