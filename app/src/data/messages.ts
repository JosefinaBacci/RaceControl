import { useCallback, useEffect } from 'react';

import { listConversations, openConversation, replyToConversation, type MailboxBox } from '@/api/messages';

import { useUnreadMessages } from './UnreadMessagesProvider';
import { useReloadOnFocus } from './useReloadOnFocus';
import { useRemote } from './useRemote';

export function useMailbox(box: MailboxBox) {
  const remote = useRemote(() => listConversations(box), [box]);
  useReloadOnFocus(remote.reload);
  return remote;
}

// Every load marks the conversation read, so the badge is refreshed after it.
export function useConversation(id: number) {
  const { reload: reloadUnread } = useUnreadMessages();
  const remote = useRemote(() => openConversation(id), [id]);
  const { state, reload } = remote;
  useReloadOnFocus(reload);

  useEffect(() => {
    if (state.status === 'ready') {
      void reloadUnread();
    }
  }, [state, reloadUnread]);

  const reply = useCallback(
    async (body: string) => {
      await replyToConversation(id, body);
      await reload();
    },
    [id, reload],
  );

  return { state, reload, reply };
}
