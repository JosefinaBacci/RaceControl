import type { FlagCode } from '@/illustrations';

export type CategoryCode = 'f1' | 'f2' | 'f3' | 'academy';

export const categories: readonly { value: CategoryCode; label: string }[] = [
  { value: 'f1', label: 'F1' },
  { value: 'f2', label: 'F2' },
  { value: 'f3', label: 'F3' },
  { value: 'academy', label: 'F1 Academy' },
];

export type Team = { id: number; category: CategoryCode; name: string; country: string; flag: FlagCode; color: string };

export const teams: readonly Team[] = [
  { id: 1, category: 'f1', name: 'Scuderia Ferrari', country: 'Italia', flag: 'it', color: '#E8002D' },
  { id: 2, category: 'f1', name: 'Red Bull Racing', country: 'Austria', flag: 'at', color: '#3671C6' },
  { id: 3, category: 'f1', name: 'Mercedes-AMG Petronas', country: 'Alemania', flag: 'de', color: '#27F4D2' },
  { id: 4, category: 'f1', name: 'McLaren', country: 'Reino Unido', flag: 'gb', color: '#FF8000' },
  { id: 5, category: 'f2', name: 'ART Grand Prix', country: 'Francia', flag: 'fr', color: '#B6BABD' },
  { id: 6, category: 'f2', name: 'PREMA Racing', country: 'Italia', flag: 'it', color: '#FF2D55' },
  { id: 7, category: 'f3', name: 'Campos Racing', country: 'España', flag: 'es', color: '#FFC300' },
  { id: 8, category: 'academy', name: 'Ferrari Driver Academy', country: 'Italia', flag: 'it', color: '#E8002D' },
];

const unassignedTeam: Team = { id: 0, category: 'f1', name: 'Escudería sin asignar', country: '—', flag: 'it', color: '#9C9CA8' };

export function teamById(teamId: number | null): Team {
  return teams.find((team) => team.id === teamId) ?? unassignedTeam;
}

export function teamName(teamId: number | null): string {
  return teamById(teamId).name;
}
