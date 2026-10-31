import type { Href } from 'expo-router';

export function controlResultsHref(eventId: number): Href {
  return { pathname: '/fia/results/controls/[id]', params: { id: String(eventId) } };
}
