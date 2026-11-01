import type { SanctionQuery } from '@/api/sanctions';

import { dateRangeQuery, type DateRangeValue } from '../history/historyFilter';
import type { SanctionStatusFilter } from './sanctionText';

// An empty driver, team or category means "any"; the ids are kept as text so
// the pickers can hold the empty choice.
export type SanctionFilterValue = DateRangeValue & {
  driverId: string;
  teamId: string;
  category: string;
  status: SanctionStatusFilter;
};

export const anyChoice = '';

export const emptySanctionFilter: SanctionFilterValue = {
  from: '',
  to: '',
  driverId: anyChoice,
  teamId: anyChoice,
  category: anyChoice,
  status: 'all',
};

export function sanctionQuery(value: SanctionFilterValue): SanctionQuery {
  return {
    ...dateRangeQuery(value),
    driverId: value.driverId ? Number(value.driverId) : undefined,
    teamId: value.teamId ? Number(value.teamId) : undefined,
    category: value.category || undefined,
    status: value.status === 'all' ? undefined : value.status,
  };
}
