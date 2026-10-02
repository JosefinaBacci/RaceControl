import { useState } from 'react';

import { ChipGroup, Screen } from '@/components';
import { useSanctions, type CategoryFilter } from '@/data/hooks';
import { categoryFilterOptions } from '@/features/filters';
import { SanctionList } from '@/features/SanctionList';

export default function PublicSanctionsScreen() {
  const [category, setCategory] = useState<CategoryFilter>('all');
  const sanctions = useSanctions(category);

  return (
    <Screen title="Sanciones" subtitle="Sanciones publicadas por la FIA a pilotos y escuderías.">
      <ChipGroup options={categoryFilterOptions} selected={category} onSelect={setCategory} accessibilityLabel="Filtrar por categoría" />
      <SanctionList title="Sanciones publicadas" sanctions={sanctions} showAcknowledgement={false} />
    </Screen>
  );
}
