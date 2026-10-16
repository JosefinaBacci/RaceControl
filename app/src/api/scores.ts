import { apiRequest, queryString } from './client';

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

export type ScoreRace = {
  id: number;
  name: string;
  gpName: string | null;
  roundNumber: number | null;
  startsAt: string;
  seasonYear: number;
  categoryCode: string;
  categoryName: string;
  circuitName: string | null;
  countryCode: string | null;
};

export type DriverResult = {
  position: number;
  driverId: number;
  firstName: string;
  lastName: string;
  countryCode: string;
  number: number | null;
  teamId: number;
  teamCode: string;
  teamName: string;
  points: number;
};

export type RaceResults = {
  race: ScoreRace;
  results: DriverResult[];
};

export type DriverStanding = DriverResult & { races: number };

export type TeamStanding = {
  position: number;
  teamId: number;
  teamCode: string;
  teamName: string;
  points: number;
};

export type Standings = {
  seasonYear: number;
  categoryCode: string;
  drivers: DriverStanding[];
  teams: TeamStanding[];
};

export type ScoreFilter = {
  season?: number;
  category?: string;
};

export type NoticeKind = 'published' | 'corrected';

export type AcknowledgementStatus = 'pending' | 'confirmed';

export type Acknowledgement = {
  status: AcknowledgementStatus;
  acknowledgedAt: string | null;
  acknowledgedBy: string | null;
};

export type ChangedPoints = {
  driverId: number;
  firstName: string;
  lastName: string;
  oldPoints: number | null;
  newPoints: number | null;
};

export type ScoreNotice = Acknowledgement & {
  id: number;
  race: ScoreRace;
  kind: NoticeKind;
  changedAt: string;
  items: ChangedPoints[];
};

export type ScoreNoticeBoard = {
  pendingCount: number;
  notices: ScoreNotice[];
};

export type TeamAcknowledgement = Acknowledgement & {
  teamId: number;
  teamCode: string;
  teamName: string;
};

export type ScoreChange = {
  id: number;
  race: ScoreRace;
  kind: NoticeKind;
  changedAt: string;
  changedBy: string;
  teams: TeamAcknowledgement[];
};

export async function listRaceResults(filter: ScoreFilter = {}): Promise<RaceResults[]> {
  const response = await apiRequest<{ races: RaceResults[] }>('GET', `/results${queryString(filter)}`);
  return response.races;
}

export function getStandings(filter: ScoreFilter): Promise<Standings> {
  return apiRequest<Standings>('GET', `/standings${queryString(filter)}`);
}

export function listTeamScoreNotices(): Promise<ScoreNoticeBoard> {
  return apiRequest<ScoreNoticeBoard>('GET', '/team/score-notices');
}

export function acknowledgeScoreChange(id: number): Promise<ScoreNotice> {
  return apiRequest<ScoreNotice>('POST', `/team/score-notices/${id}/acknowledgement`);
}

export async function listScoreChanges(): Promise<ScoreChange[]> {
  const response = await apiRequest<{ changes: ScoreChange[] }>('GET', '/score-changes');
  return response.changes;
}
