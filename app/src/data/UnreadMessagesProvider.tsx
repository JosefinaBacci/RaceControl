import { createContext, useContext, type ReactNode } from 'react';

import { listConversations } from '@/api/messages';

import { useRemote } from './useRemote';

type UnreadMessages = {
  unreadCount: number;
  reload: () => Promise<void>;
};

const UnreadMessagesContext = createContext<UnreadMessages | null>(null);

// The tab badge and the conversation screen share the count, so opening a
// conversation clears its unread messages from the badge at once.
export function UnreadMessagesProvider({ children }: { children: ReactNode }) {
  const { state, reload } = useRemote(() => listConversations('inbox'), []);
  const unreadCount = state.status === 'ready' ? state.data.unreadCount : 0;

  return <UnreadMessagesContext.Provider value={{ unreadCount, reload }}>{children}</UnreadMessagesContext.Provider>;
}

export function useUnreadMessages(): UnreadMessages {
  const unread = useContext(UnreadMessagesContext);
  if (!unread) {
    throw new Error('useUnreadMessages must be used inside UnreadMessagesProvider');
  }
  return unread;
}
