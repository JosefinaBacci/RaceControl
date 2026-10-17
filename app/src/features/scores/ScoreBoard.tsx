import { useState } from 'react';
import { StyleSheet, View } from 'react-native';

import type { RaceResults, Standings } from '@/api/scores';
import { AppText, ChipGroup, HeroCard, SelectField } from '@/components';
import { useRaceResults, useSelectedRace, useStandings } from '@/data/scores';
import { formatEventDate, formatPoints } from '@/format';
import { RemoteContent } from '@/features/RemoteContent';
import { StandingsTable } from '@/features/StandingsTable';
import { categories, categoryLabel } from '@/mocks/catalog';
import { categoryPhoto } from '@/photos';
import { spacing } from '@/theme';

import { driverRows, teamRows } from './standingsRows';

// The championship of one category: its accumulated standings and the results
// of each raced race, picked with the race filter.
export function ScoreBoard({ initialCategory }: { initialCategory: string }) {
  const [category, setCategory] = useState(initialCategory);
  const standings = useStandings(category);
  const results = useRaceResults(category);

  return (
    <>
      <ChipGroup<string> variant="underline" options={categories} selected={category} onSelect={setCategory} accessibilityLabel="Elegir categoría" />
      <RemoteContent state={standings.state} onRetry={standings.reload}>
        {(data) => <ChampionshipStandings standings={data} category={category} />}
      </RemoteContent>
      <RemoteContent state={results.state} onRetry={results.reload}>
        {(races) => <RaceResultsSection races={races} />}
      </RemoteContent>
    </>
  );
}

function ChampionshipStandings({ standings, category }: { standings: Standings; category: string }) {
  const leader = standings.drivers[0];

  return (
    <>
      <HeroCard
        overline={`Temporada ${standings.seasonYear}`}
        title={`Campeonato de ${categoryLabel(category)}`}
        subtitle={leader ? `Líder: ${leader.firstName} ${leader.lastName} · ${formatPoints(leader.points)} pts` : 'Todavía no hay puntajes cargados.'}
        photo={categoryPhoto(category)}
      />
      <View style={styles.columns}>
        <View style={styles.column}>
          <StandingsTable rows={driverRows(standings.drivers)} />
        </View>
        <View style={styles.column}>
          <StandingsTable title="Clasificación de escuderías" rows={teamRows(standings.teams)} nameLabel="Escudería" detailLabel="" />
        </View>
      </View>
    </>
  );
}

function RaceResultsSection({ races }: { races: RaceResults[] }) {
  const { selected, select } = useSelectedRace(races);

  if (!selected) {
    return <AppText color="textMuted">Todavía no hay resultados publicados en esta categoría.</AppText>;
  }
  const options = races.map(({ race }) => ({ value: String(race.id), label: race.gpName ?? race.name }));

  return (
    <View style={styles.section}>
      <AppText variant="heading">Resultados por carrera</AppText>
      <SelectField options={options} selected={String(selected.race.id)} onSelect={select} accessibilityLabel="Elegir carrera" />
      <StandingsTable
        title={selected.race.name}
        action={
          <AppText variant="caption" color="textMuted">
            {formatEventDate(selected.race.startsAt)}
          </AppText>
        }
        rows={driverRows(selected.results)}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  columns: { flexDirection: 'row', flexWrap: 'wrap', gap: spacing.lg },
  column: { flexGrow: 1, flexBasis: 360, gap: spacing.lg },
  section: { gap: spacing.md },
});
