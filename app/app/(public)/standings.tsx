import { useState } from 'react';

import { ChipGroup, Screen } from '@/components';
import { useStandings } from '@/data/hooks';
import { StandingsTable } from '@/features/StandingsTable';
import { categories, type CategoryCode } from '@/mocks/catalog';

export default function PublicStandingsScreen() {
  const [category, setCategory] = useState<CategoryCode>('f1');
  const rows = useStandings(category);

  return (
    <Screen title="Puntajes" subtitle="Clasificación del campeonato por categoría.">
      <ChipGroup variant="underline" options={categories} selected={category} onSelect={setCategory} accessibilityLabel="Elegir categoría" />
      <StandingsTable rows={rows} />
    </Screen>
  );
}
