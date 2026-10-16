import { listTeams, type TeamSummary } from '@/api/teams';

import { useRemote } from './useRemote';

// The team's championship comes from the catalogue the API serves, never from the
// mocks, so a team of any category finds its own standings.
export async function findTeam(teamId: number | null): Promise<TeamSummary | null> {
  const teams = await listTeams();
  return teams.find((team) => team.id === teamId) ?? null;
}

export function useTeam(teamId: number | null) {
  return useRemote(() => findTeam(teamId), [teamId]);
}
