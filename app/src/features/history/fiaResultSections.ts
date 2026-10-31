import type { ComponentType } from 'react';

import { FiaControlsCard } from '@/features/technicalControls/FiaControlsCard';
import { FiaTyreTestsCard } from '@/features/tyreTests/FiaTyreTestsCard';

// The FIA "Resultados" tab lists races first and then one card per event kind
// with official results. A new kind adds its card here instead of editing the tab.
export const fiaResultSections: readonly { key: string; Component: ComponentType }[] = [
  { key: 'controls', Component: FiaControlsCard },
  { key: 'tyres', Component: FiaTyreTestsCard },
];
