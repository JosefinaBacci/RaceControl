// The shape every official-results history shares (technical controls, tyre
// tests), so the same list, badges and chain render any of them. Days travel as
// YYYY-MM-DD in UTC; see features/history/day.ts.

export type HistoryType = {
  code: string;
  name: string;
};

export type HistoryEntry = {
  id: number;
  eventId: number;
  conductedOn: string;
  type: HistoryType;
  passed: boolean;
  notes: string;
  isCorrection: boolean;
  correctionNote: string | null;
  seasonYear: number;
  categoryCode: string;
  recordedAt: string;
};

// The FIA-only view of an entry: who recorded it and which entry it replaced.
export type ResultRevision = HistoryEntry & {
  supersedesId: number | null;
  recordedBy: { username: string };
};

export type HistoryQuery = {
  from?: string;
  to?: string;
  type?: string;
};

export type ResultInput = {
  conductedOn: string;
  passed: boolean;
  notes: string;
};

export type ResultCorrectionInput = ResultInput & {
  correctionNote: string;
};

export type HistoryTypeList = {
  types: HistoryType[];
};
