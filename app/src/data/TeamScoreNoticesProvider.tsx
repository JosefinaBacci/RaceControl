import { createContext, useCallback, useContext, type ReactNode } from 'react';

import { acknowledgeScoreChange, listTeamScoreNotices, type ScoreNoticeBoard } from '@/api/scores';

import { useRemote, type RemoteState } from './useRemote';

type TeamScoreNotices = {
  state: RemoteState<ScoreNoticeBoard>;
  reload: () => Promise<void>;
  acknowledge: (changeLogId: number) => Promise<void>;
};

const TeamScoreNoticesContext = createContext<TeamScoreNotices | null>(null);

// The tab badge, the dashboard and the notices screen read the same board, so a
// confirmation shows everywhere at once.
export function TeamScoreNoticesProvider({ children }: { children: ReactNode }) {
  const { state, reload } = useRemote(listTeamScoreNotices, []);

  const acknowledge = useCallback(
    async (changeLogId: number) => {
      await acknowledgeScoreChange(changeLogId);
      await reload();
    },
    [reload],
  );

  return <TeamScoreNoticesContext.Provider value={{ state, reload, acknowledge }}>{children}</TeamScoreNoticesContext.Provider>;
}

export function useTeamScoreNotices(): TeamScoreNotices {
  const notices = useContext(TeamScoreNoticesContext);
  if (!notices) {
    throw new Error('useTeamScoreNotices must be used inside TeamScoreNoticesProvider');
  }
  return notices;
}

export function usePendingScoreNotices(): number {
  const { state } = useTeamScoreNotices();
  return state.status === 'ready' ? state.data.pendingCount : 0;
}
