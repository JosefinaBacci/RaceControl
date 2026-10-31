import { apiRequest, queryString } from './client';
import type {
  HistoryEntry,
  HistoryQuery,
  HistoryType,
  HistoryTypeList,
  ResultCorrectionInput,
  ResultInput,
  ResultRevision,
} from './history';

export type GradedTeam = {
  id: number;
  name: string;
};

export type ControlResult = HistoryEntry & { team: GradedTeam };

export type ControlRevision = ResultRevision & { team: GradedTeam };

export type ControlHistoryQuery = HistoryQuery & {
  category?: string;
  team?: number;
};

export type TechnicalControl = {
  eventId: number;
  controlType: string;
  startsAt: string;
  endsAt: string;
  isActive: boolean;
  seasonYear: number;
  categoryCode: string;
  categoryName: string;
};

export type ControlSheetTeam = GradedTeam & {
  isActive: boolean;
  revisions: ControlRevision[];
};

export type ControlSheet = {
  control: TechnicalControl;
  teams: ControlSheetTeam[];
};

export async function listControlHistory(query: ControlHistoryQuery): Promise<ControlResult[]> {
  const response = await apiRequest<{ results: ControlResult[] }>('GET', `/technical-controls/results${queryString(query)}`);
  return response.results;
}

export async function listControlTypes(): Promise<HistoryType[]> {
  const response = await apiRequest<HistoryTypeList>('GET', '/technical-controls/types');
  return response.types;
}

export function getControlSheet(eventId: number): Promise<ControlSheet> {
  return apiRequest<ControlSheet>('GET', `/technical-controls/${eventId}/results`);
}

export function recordControlResult(eventId: number, teamId: number, input: ResultInput): Promise<ControlRevision> {
  return apiRequest<ControlRevision>('POST', `/technical-controls/${eventId}/results`, { ...input, teamId });
}

export function correctControlResult(resultId: number, input: ResultCorrectionInput): Promise<ControlRevision> {
  return apiRequest<ControlRevision>('POST', `/technical-controls/results/${resultId}/corrections`, input);
}
