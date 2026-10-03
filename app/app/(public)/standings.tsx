import { useState } from 'react';

import { ChipGroup, HeroCard, Screen } from '@/components';
import { useStandings } from '@/data/hooks';
import { StandingsTable } from '@/features/StandingsTable';
import { categories, categoryLabel, type CategoryCode } from '@/mocks/catalog';
import { categoryPhotos } from '@/photos';

export default function PublicStandingsScreen() {
  const [category, setCategory] = useState<CategoryCode>('f1');
  const rows = useStandings(category);
  const label = categoryLabel(category);

  return (
    <Screen title="Puntajes" subtitle="Clasificación del campeonato por categoría.">
      <ChipGroup variant="underline" options={categories} selected={category} onSelect={setCategory} accessibilityLabel="Elegir categoría" />
      <HeroCard
        overline="Temporada 2026"
        title={`Campeonato de ${label}`}
        subtitle={rows[0] ? `Líder: ${rows[0].driver} · ${rows[0].points} pts` : undefined}
        photo={categoryPhotos[category]}
      />
      <StandingsTable rows={rows} />
    </Screen>
  );
}
