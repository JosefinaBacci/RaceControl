import { useMemo, useState } from 'react';

import type { CategoryCode, CategoryFilter } from '@/mocks/catalog';
import { sanctions, scoreNotices } from '@/mocks/data';

const matchesCategory = (category: CategoryFilter) => (item: { category: CategoryCode }) =>
  category === 'all' || item.category === category;

export function useSanctions(category: CategoryFilter) {
  return useMemo(() => sanctions.filter(matchesCategory(category)), [category]);
}

export function useScoreNotices() {
  return scoreNotices;
}

export type InboxItem = {
  key: string;
  title: string;
  detail: string;
  acknowledged: boolean;
};

export function useTeamInbox(teamId: number | null) {
  const [acknowledgedKeys, setAcknowledgedKeys] = useState<ReadonlySet<string>>(new Set());

  const items = useMemo<InboxItem[]>(
    () =>
      sanctions
        .filter((sanction) => sanction.teamId === teamId)
        .map((sanction) => {
          const key = `sanction-${sanction.id}`;
          return {
            key,
            title: sanction.penalty,
            detail: `${sanction.subject} · ${sanction.event}`,
            acknowledged: sanction.acknowledged || acknowledgedKeys.has(key),
          };
        }),
    [teamId, acknowledgedKeys],
  );

  const acknowledge = (key: string) => setAcknowledgedKeys((previous) => new Set(previous).add(key));
  const pendingCount = countPending(items);

  return { items, pendingCount, acknowledge };
}

export function usePendingSanctions() {
  return useMemo(() => countPending(sanctions), []);
}

function countPending(items: readonly { acknowledged: boolean }[]): number {
  return items.filter((item) => !item.acknowledged).length;
}
