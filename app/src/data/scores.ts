import { getScoreGrid } from '@/api/scores';

import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useScoreGrid(eventId: number) {
  const remote = useRemote(() => getScoreGrid(eventId), [eventId]);
  useReloadOnFocus(remote.reload);
  return remote;
}
