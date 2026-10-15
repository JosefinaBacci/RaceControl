import { apiRequest } from './client';

// What both the team screen and the public directory show of a driver.
export type DriverSummary = {
  id: number;
  firstName: string;
  lastName: string;
  countryCode: string;
  isStarter: boolean;
  number: number | null;
};

export type TeamDriver = DriverSummary & {
  licenseNumber: string;
};

export type PublicDriver = DriverSummary & {
  teamId: number;
  teamName: string;
  categoryCode: string;
  categoryName: string;
};

export type DriverFilter = {
  category?: string;
  search?: string;
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

export async function listDrivers(filter: DriverFilter = {}): Promise<PublicDriver[]> {
  const response = await apiRequest<{ drivers: PublicDriver[] }>('GET', `/drivers${queryString(filter)}`);
  return response.drivers;
}

export function getDriver(id: number): Promise<PublicDriver> {
  return apiRequest<PublicDriver>('GET', `/drivers/${id}`);
}

function queryString(filter: DriverFilter): string {
  const params = { category: filter.category, q: filter.search };
  const pairs = Object.entries(params)
    .filter((entry): entry is [string, string] => Boolean(entry[1]))
    .map(([key, value]) => `${key}=${encodeURIComponent(value)}`);
  return pairs.length > 0 ? `?${pairs.join('&')}` : '';
}
