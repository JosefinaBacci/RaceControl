import { apiRequest } from './client';

export type TeamSummary = {
  id: number;
  code: string;
  name: string;
  categoryCode: string;
  categoryName: string;
};

export async function listTeams(): Promise<TeamSummary[]> {
  const response = await apiRequest<{ teams: TeamSummary[] }>('GET', '/teams');
  return response.teams;
}
