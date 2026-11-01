import { usePendingSanctionNotices } from '@/data/TeamSanctionNoticesProvider';
import { usePendingScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { useUnreadMessages } from '@/data/UnreadMessagesProvider';

import type { TabBadges } from './tabSpecs';

export function useTeamTabBadges(): TabBadges {
  return {
    inbox: usePendingSanctionNotices(),
    notifications: usePendingScoreNotices(),
    messages: useUnreadMessages().unreadCount,
  };
}

export function useFiaTabBadges(): TabBadges {
  return { messages: useUnreadMessages().unreadCount };
}
