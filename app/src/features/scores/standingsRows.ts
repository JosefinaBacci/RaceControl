import type { DriverResult, TeamStanding } from '@/api/scores';
import type { StandingsRow } from '@/features/StandingsTable';
import { teamById } from '@/mocks/catalog';

export function driverRows(drivers: readonly DriverResult[]): StandingsRow[] {
  return drivers.map((driver) => ({
    key: `driver-${driver.driverId}`,
    position: driver.position,
    name: `${driver.firstName} ${driver.lastName}`,
    color: teamById(driver.teamId).color,
    detail: driver.teamName,
    points: driver.points,
  }));
}

export function teamRows(teams: readonly TeamStanding[]): StandingsRow[] {
  return teams.map((team) => ({
    key: `team-${team.teamId}`,
    position: team.position,
    name: team.teamName,
    color: teamById(team.teamId).color,
    points: team.points,
  }));
}
