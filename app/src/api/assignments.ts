import { apiRequest, queryString } from './client';
export type Assignment = {
 id: number; driverId: number; firstName: string; lastName: string;
 teamId: number | null; teamName: string; previousTeamName: string;
 isStarter: boolean; effectiveFrom: string; effectiveUntil: string | null;
};
export async function listAssignments(teamId?: number): Promise<Assignment[]> {
 const response = await apiRequest<{ assignments: Assignment[] }>('GET', `/assignments${queryString({ teamId })}`);
 return response.assignments;
}
