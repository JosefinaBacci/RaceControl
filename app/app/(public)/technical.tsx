import { useState } from 'react';

import { ChipGroup, Notice, Screen } from '@/components';
import { technicalSections } from '@/features/history/technicalSections';

export default function PublicTechnicalScreen() {
  const [selected, setSelected] = useState(technicalSections[0]?.value ?? '');
  const section = technicalSections.find((candidate) => candidate.value === selected);

  return (
    <Screen title="Técnica" subtitle="Resultados oficiales de los controles técnicos y las pruebas de neumáticos.">
      <ChipGroup variant="underline" options={technicalSections} selected={selected} onSelect={setSelected} accessibilityLabel="Sección" />
      {section ? <section.Component /> : <Notice tone="info" message="Todavía no hay resultados técnicos publicados." />}
    </Screen>
  );
}
