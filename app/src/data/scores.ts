import { useState } from 'react';

import { getStandings, listRaceResults, listScoreChanges, type RaceResults, type ScoreChange } from '@/api/scores';

import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useStandings(category: string) {
  const remote = useRemote(() => getStandings({ category }), [category]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useTeamPoints(teamId: number | null, category: string): number | null {
  const { state } = useStandings(category);
  if (state.status !== 'ready') {
    return null;
  }
  return state.data.teams.find((team) => team.teamId === teamId)?.points ?? 0;
}

export function useRaceResults(category: string) {
  const remote = useRemote(() => listRaceResults({ category }), [category]);
  useReloadOnFocus(remote.reload);
  return remote;
}

export function useScoreChanges() {
  const remote = useRemote(listScoreChanges, []);
  useReloadOnFocus(remote.reload);
  return remote;
}

// The FIA follows every affected team of every change, so a team that has yet to
// confirm two corrections counts twice.
export function usePendingScoreConfirmations(): number {
  const { state } = useScoreChanges();
  return state.status === 'ready' ? countPendingConfirmations(state.data) : 0;
}

function countPendingConfirmations(changes: readonly ScoreChange[]): number {
  return changes.flatMap((change) => change.teams).filter((team) => team.status === 'pending').length;
}

// The race filter starts on the latest raced race and falls back to it whenever
// the chosen one leaves the list, as when the category changes.
export function useSelectedRace(races: readonly RaceResults[]) {
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const selected = races.find((entry) => String(entry.race.id) === selectedId) ?? races[races.length - 1] ?? null;
  return { selected, select: setSelectedId };
}
