import { apiRequest } from './client';

export type ScoreGridEntry = {
  driverId: number;
  firstName: string;
  lastName: string;
  countryCode: string;
  teamId: number;
  teamName: string;
  number: number | null;
  points: number | null;
};

export type ScoreGrid = {
  eventId: number;
  categoryCode: string;
  categoryName: string;
  entries: ScoreGridEntry[];
};

export type ScoreEntryInput = {
  driverId: number;
  points: number;
};

export function getScoreGrid(eventId: number): Promise<ScoreGrid> {
  return apiRequest<ScoreGrid>('GET', `/events/${eventId}/scores`);
}

export function replaceScores(eventId: number, entries: ScoreEntryInput[]): Promise<ScoreGrid> {
  return apiRequest<ScoreGrid>('PUT', `/events/${eventId}/scores`, { entries });
}
