import type { ComponentType } from 'react';

import { ControlHistory } from '@/features/technicalControls/ControlHistory';

// The public "Técnica" tab: one section per official-results history. A new
// history adds its line here instead of editing the tab.
export type TechnicalSection = {
  value: string;
  label: string;
  Component: ComponentType;
};

export const technicalSections: readonly TechnicalSection[] = [
  { value: 'controls', label: 'Controles técnicos', Component: ControlHistory },
];
