import { useMemo, useState } from 'react';

import type { CategoryCode, CategoryFilter } from '@/mocks/catalog';
import { driverStandings, notifications, sanctions, scoreNotices } from '@/mocks/data';

const matchesCategory = (category: CategoryFilter) => (item: { category: CategoryCode }) =>
  category === 'all' || item.category === category;

export function useStandings(category: CategoryCode, limit?: number) {
  return useMemo(() => driverStandings[category].slice(0, limit), [category, limit]);
}

export function useSanctions(category: CategoryFilter) {
  return useMemo(() => sanctions.filter(matchesCategory(category)), [category]);
}

export function useTeamPoints(teamId: number | null) {
  return useMemo(
    () =>
      Object.values(driverStandings)
        .flat()
        .filter((row) => row.teamId === teamId)
        .reduce((total, row) => total + row.points, 0),
    [teamId],
  );
}

export function useNotifications() {
  return notifications;
}

export function useScoreNotices() {
  return scoreNotices;
}

export type InboxItem = {
  key: string;
  kind: 'Sanción' | 'Puntaje';
  title: string;
  detail: string;
  acknowledged: boolean;
};

export function useTeamInbox(teamId: number | null) {
  const [acknowledgedKeys, setAcknowledgedKeys] = useState<ReadonlySet<string>>(new Set());

  const items = useMemo<InboxItem[]>(() => {
    const sanctionItems = sanctions
      .filter((sanction) => sanction.teamId === teamId)
      .map((sanction) => ({
        key: `sanction-${sanction.id}`,
        kind: 'Sanción' as const,
        title: sanction.penalty,
        detail: `${sanction.subject} · ${sanction.event}`,
        acknowledged: sanction.acknowledged,
      }));
    const scoreItems = scoreNotices
      .filter((notice) => notice.teamId === teamId)
      .map((notice) => ({
        key: `score-${notice.id}`,
        kind: 'Puntaje' as const,
        title: `${notice.points} puntos`,
        detail: notice.event,
        acknowledged: notice.acknowledged,
      }));
    return [...sanctionItems, ...scoreItems].map((item) => ({
      ...item,
      acknowledged: item.acknowledged || acknowledgedKeys.has(item.key),
    }));
  }, [teamId, acknowledgedKeys]);

  const acknowledge = (key: string) => setAcknowledgedKeys((previous) => new Set(previous).add(key));
  const pendingCount = countPending(items);

  return { items, pendingCount, acknowledge };
}

export function usePendingAcknowledgements() {
  return useMemo(() => ({ sanctions: countPending(sanctions), scores: countPending(scoreNotices) }), []);
}

function countPending(items: readonly { acknowledged: boolean }[]): number {
  return items.filter((item) => !item.acknowledged).length;
}
