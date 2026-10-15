import { apiRequest } from './client';

export type TeamDriver = {
  id: number;
  firstName: string;
  lastName: string;
  licenseNumber: string;
  countryCode: string;
  isStarter: boolean;
  number: number | null;
};

export type NewDriver = {
  firstName: string;
  lastName: string;
  licenseNumber: string;
  countryCode: string;
  isStarter: boolean;
  number: number | null;
};

export type DriverChanges = {
  firstName?: string;
  lastName?: string;
  licenseNumber?: string;
  countryCode?: string;
  isStarter?: boolean;
  number?: number;
};

export async function listTeamDrivers(): Promise<TeamDriver[]> {
  const response = await apiRequest<{ drivers: TeamDriver[] }>('GET', '/team/drivers');
  return response.drivers;
}

export function getTeamDriver(id: number): Promise<TeamDriver> {
  return apiRequest<TeamDriver>('GET', `/team/drivers/${id}`);
}

export function createDriver(input: NewDriver): Promise<TeamDriver> {
  return apiRequest<TeamDriver>('POST', '/team/drivers', input);
}

export function updateDriver(id: number, changes: DriverChanges): Promise<TeamDriver> {
  return apiRequest<TeamDriver>('PATCH', `/team/drivers/${id}`, changes);
}

export function releaseDriver(id: number): Promise<void> {
  return apiRequest<void>('POST', `/team/drivers/${id}/release`);
}
