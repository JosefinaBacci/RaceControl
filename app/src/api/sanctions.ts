import { apiRequest, queryString } from './client';
import type { Acknowledgement } from './scores';

// A sanction is never rewritten: issuing, amending and revoking each add a
// version, and the latest one is `current`. The public shapes carry no author;
// only the FIA follow-up names who changed what.

export type SanctionAction = 'issued' | 'amended' | 'revoked';

export type SanctionStatus = 'active' | 'revoked';

export type PenaltyUnit = 'seconds' | 'grid_places' | 'license_points' | 'euros' | 'races';

export type PenaltyType = {
  code: string;
  name: string;
  unit: PenaltyUnit | null;
  maxValue: number | null;
};

export type Penalty = {
  code: string;
  name: string;
  unit: PenaltyUnit | null;
  value: number | null;
};

export type SanctionSubject = {
  kind: 'driver' | 'team';
  driverId: number | null;
  driverFirstName: string | null;
  driverLastName: string | null;
};

export type SanctionTeam = {
  id: number;
  code: string;
  name: string;
};

export type SanctionVersion = {
  id: number;
  action: SanctionAction;
  penalty: Penalty;
  reason: string;
  event: { id: number; name: string; startsAt: string } | null;
  note: string | null;
  changedAt: string;
};

export type Sanction = {
  id: number;
  status: SanctionStatus;
  subject: SanctionSubject;
  team: SanctionTeam;
  categoryCode: string;
  categoryName: string;
  issuedAt: string;
  current: SanctionVersion;
};

export type SanctionDetail = Sanction & {
  versions: SanctionVersion[];
};

export type SanctionQuery = {
  driverId?: number;
  teamId?: number;
  category?: string;
  status?: SanctionStatus | 'all';
  from?: string;
  to?: string;
};

export type SanctionTerms = {
  penaltyType: string;
  penaltyValue: number | null;
  reason: string;
  eventId: number | null;
};

export type NewSanction = SanctionTerms & {
  driverId?: number;
  teamId?: number;
};

export type SanctionAmendment = SanctionTerms & {
  note: string;
};

export type SanctionNotice = Acknowledgement & {
  id: number;
  sanction: Sanction;
  isOwnTeam: boolean;
};

export type SanctionNoticeBoard = {
  pendingCount: number;
  notices: SanctionNotice[];
};

export type SanctionRecipient = Acknowledgement & {
  team: SanctionTeam;
};

export type SanctionVersionFollowUp = SanctionVersion & {
  changedBy: string;
  isCurrent: boolean;
  confirmedCount: number;
  teams: SanctionRecipient[];
};

export type SanctionFollowUp = {
  sanction: Sanction;
  versions: SanctionVersionFollowUp[];
};

export async function listPenaltyTypes(): Promise<PenaltyType[]> {
  const response = await apiRequest<{ penaltyTypes: PenaltyType[] }>('GET', '/penalty-types');
  return response.penaltyTypes;
}

export async function listSanctions(query: SanctionQuery = {}): Promise<Sanction[]> {
  const response = await apiRequest<{ sanctions: Sanction[] }>('GET', `/sanctions${queryString(query)}`);
  return response.sanctions;
}

export function getSanction(id: number): Promise<SanctionDetail> {
  return apiRequest<SanctionDetail>('GET', `/sanctions/${id}`);
}

export function issueSanction(input: NewSanction): Promise<SanctionDetail> {
  return apiRequest<SanctionDetail>('POST', '/sanctions', input);
}

export function amendSanction(id: number, amendment: SanctionAmendment): Promise<SanctionDetail> {
  return apiRequest<SanctionDetail>('PUT', `/sanctions/${id}`, amendment);
}

export function revokeSanction(id: number, note: string): Promise<SanctionDetail> {
  return apiRequest<SanctionDetail>('POST', `/sanctions/${id}/revocation`, { note });
}

export function listTeamSanctionNotices(): Promise<SanctionNoticeBoard> {
  return apiRequest<SanctionNoticeBoard>('GET', '/team/sanction-notices');
}

export function acknowledgeSanctionVersion(versionId: number): Promise<SanctionNotice> {
  return apiRequest<SanctionNotice>('POST', `/team/sanction-notices/${versionId}/acknowledgement`);
}

export async function listSanctionFollowUps(): Promise<SanctionFollowUp[]> {
  const response = await apiRequest<{ followUps: SanctionFollowUp[] }>('GET', '/sanction-acknowledgements');
  return response.followUps;
}
