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

// A tyre test result's type is the compound it evaluated.
export type TyreTestResult = HistoryEntry & {
  circuitName: string;
};

export type TyreTestRevision = ResultRevision & {
  circuitName: string;
  isCurrent: boolean;
};

export type TyreTestResultInput = ResultInput & {
  compoundCode: string;
};

export async function getTyreCompounds(): Promise<HistoryType[]> {
  const list = await apiRequest<HistoryTypeList>('GET', '/tyre-tests/types');
  return list.types;
}

export async function getTyreTestHistory(query: HistoryQuery): Promise<TyreTestResult[]> {
  const list = await apiRequest<{ results: TyreTestResult[] }>('GET', `/tyre-tests/results${queryString(query)}`);
  return list.results;
}

export async function getTyreTestRevisions(eventId: number): Promise<TyreTestRevision[]> {
  const list = await apiRequest<{ results: TyreTestRevision[] }>('GET', `/tyre-tests/${eventId}/results`);
  return list.results;
}

export function recordTyreTestResult(eventId: number, input: TyreTestResultInput): Promise<TyreTestRevision> {
  return apiRequest<TyreTestRevision>('POST', `/tyre-tests/${eventId}/results`, input);
}

export function correctTyreTestResult(resultId: number, input: ResultCorrectionInput): Promise<TyreTestRevision> {
  return apiRequest<TyreTestRevision>('POST', `/tyre-tests/results/${resultId}/corrections`, input);
}
