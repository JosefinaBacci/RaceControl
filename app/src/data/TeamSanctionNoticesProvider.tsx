import { createContext, useCallback, useContext, type ReactNode } from 'react';

import { acknowledgeSanctionVersion, listTeamSanctionNotices, type SanctionNoticeBoard } from '@/api/sanctions';

import { useRemote, type RemoteState } from './useRemote';

type TeamSanctionNotices = {
  state: RemoteState<SanctionNoticeBoard>;
  reload: () => Promise<void>;
  acknowledge: (versionId: number) => Promise<void>;
};

const TeamSanctionNoticesContext = createContext<TeamSanctionNotices | null>(null);

// The tab badge, the dashboard and the inbox read the same board, so a
// confirmation shows everywhere at once.
export function TeamSanctionNoticesProvider({ children }: { children: ReactNode }) {
  const { state, reload } = useRemote(listTeamSanctionNotices, []);

  const acknowledge = useCallback(
    async (versionId: number) => {
      try {
        await acknowledgeSanctionVersion(versionId);
      } finally {
        await reload();
      }
    },
    [reload],
  );

  return <TeamSanctionNoticesContext.Provider value={{ state, reload, acknowledge }}>{children}</TeamSanctionNoticesContext.Provider>;
}

export function useTeamSanctionNotices(): TeamSanctionNotices {
  const notices = useContext(TeamSanctionNoticesContext);
  if (!notices) {
    throw new Error('useTeamSanctionNotices must be used inside TeamSanctionNoticesProvider');
  }
  return notices;
}

export function usePendingSanctionNotices(): number {
  const { state } = useTeamSanctionNotices();
  return state.status === 'ready' ? state.data.pendingCount : 0;
}
