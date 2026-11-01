import type { Sanction, SanctionQuery } from '@/api/sanctions';

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

type Choice = { value: string; label: string };

export function sanctionSubjects(history: readonly Sanction[]) {
  const drivers = new Map<number, string>();
  const teams = new Map<number, string>();
  for (const { subject, team } of history) {
    teams.set(team.id, team.name);
    if (subject.driverId !== null) {
      drivers.set(subject.driverId, [subject.driverFirstName, subject.driverLastName].filter(Boolean).join(' '));
    }
  }
  return { drivers: sortedChoices(drivers), teams: sortedChoices(teams) };
}

function sortedChoices(names: Map<number, string>): Choice[] {
  return Array.from(names, ([id, label]) => ({ value: String(id), label }))
    .sort((a, b) => a.label.localeCompare(b.label, 'es'));
}
