import { listEvents } from '@/api/calendar';
import { listDrivers, type PublicDriver } from '@/api/drivers';
import {
  getSanction,
  listPenaltyTypes,
  listSanctionFollowUps,
  listSanctions,
  type PenaltyType,
  type SanctionFollowUp,
  type SanctionQuery,
} from '@/api/sanctions';
import { listTeams, type TeamSummary } from '@/api/teams';

import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useSanctions(query: SanctionQuery) {
  const remote = useRemote(() => listSanctions(query), [
    query.driverId,
    query.teamId,
    query.category,
    query.status,
    query.from,
    query.to,
  ]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useSanction(id: number) {
  const remote = useRemote(() => getSanction(id), [id]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export type SanctionCatalogs = {
  penaltyTypes: PenaltyType[];
  drivers: PublicDriver[];
  teams: TeamSummary[];
};

export function useSanctionCatalogs() {
  return useRemote(async (): Promise<SanctionCatalogs> => {
    const [penaltyTypes, drivers, teams] = await Promise.all([listPenaltyTypes(), listDrivers(), listTeams()]);
    return { penaltyTypes, drivers, teams };
  }, []);
}

export function useChampionshipEvents(category: string | null) {
  return useRemote(async () => {
    if (category === null) {
      return [];
    }
    const events = await listEvents({ category });
    return events.filter((event) => event.isActive);
  }, [category]);
}

export function useSanctionFollowUps() {
  const remote = useRemote(listSanctionFollowUps, []);
  useReloadOnFocus(remote.reload);
  return remote;
}

// A sanction is pending while some team has not confirmed the version that stands.
export function hasPendingConfirmation(followUp: SanctionFollowUp): boolean {
  const current = followUp.versions.find((version) => version.isCurrent);
  return current !== undefined && current.confirmedCount < current.teams.length;
}

export function usePendingSanctionConfirmations(): number {
  const { state } = useSanctionFollowUps();
  return state.status === 'ready' ? state.data.filter(hasPendingConfirmation).length : 0;
}
