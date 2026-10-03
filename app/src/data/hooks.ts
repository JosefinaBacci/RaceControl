import { useMemo, useState } from 'react';

import type { Role } from '@/api/auth';
import type { CategoryCode } from '@/mocks/catalog';
import {
  calendarEvents,
  circuitFacts,
  driverStandings,
  drivers,
  managedUsers,
  notifications,
  sanctions,
  scoreNotices,
  type CalendarEvent,
} from '@/mocks/data';

export type CategoryFilter = CategoryCode | 'all';
export type EventPeriod = 'upcoming' | 'past';

const byDate = (a: CalendarEvent, b: CalendarEvent) => a.date.localeCompare(b.date);
const matchesCategory = (category: CategoryFilter) => (item: { category: CategoryCode }) =>
  category === 'all' || item.category === category;

export function useCalendar(category: CategoryFilter, period: EventPeriod = 'upcoming') {
  return useMemo(() => {
    const wanted = period === 'upcoming' ? 'scheduled' : 'finished';
    const events = calendarEvents.filter(matchesCategory(category)).filter((event) => event.status === wanted).sort(byDate);
    return period === 'past' ? events.reverse() : events;
  }, [category, period]);
}

export function useUpcomingEvents(limit: number) {
  return useMemo(() => calendarEvents.filter((event) => event.status === 'scheduled').sort(byDate).slice(0, limit), [limit]);
}

export function useNextRace() {
  return useMemo(() => {
    const race = calendarEvents.filter((event) => event.kind === 'race' && event.status === 'scheduled').sort(byDate)[0];
    const [year, month, day] = race.date.split('-').map(Number);
    const [hour, minute] = (race.startTime ?? '00:00').split(':').map(Number);
    return {
      race,
      startsAt: new Date(year, month - 1, day, hour, minute),
      facts: circuitFacts[race.circuit],
    };
  }, []);
}

export function useStandings(category: CategoryCode, limit?: number) {
  return useMemo(() => driverStandings[category].slice(0, limit), [category, limit]);
}

export function useSanctions(category: CategoryFilter) {
  return useMemo(() => sanctions.filter(matchesCategory(category)), [category]);
}

export function useTeamDrivers(teamId: number | null) {
  return useMemo(() => drivers.filter((driver) => driver.teamId === teamId), [teamId]);
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

export function useActiveAccountCount() {
  return useMemo(() => managedUsers.filter((user) => user.isActive).length, []);
}

function countPending(items: readonly { acknowledged: boolean }[]): number {
  return items.filter((item) => !item.acknowledged).length;
}

export type RoleFilter = Role | 'all';

export function useManagedUsers(search: string, role: RoleFilter) {
  return useMemo(() => {
    const needle = search.trim().toLowerCase();
    return managedUsers
      .filter((user) => role === 'all' || user.role === role)
      .filter((user) => needle === '' || user.username.includes(needle) || user.email.includes(needle));
  }, [search, role]);
}
