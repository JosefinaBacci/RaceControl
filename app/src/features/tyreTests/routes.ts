import type { Href } from 'expo-router';

export function tyreTestResultsHref(eventId: number): Href {
  return { pathname: '/fia/results/tyres/[id]', params: { id: String(eventId) } };
}
