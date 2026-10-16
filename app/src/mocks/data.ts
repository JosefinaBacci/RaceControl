import type { CategoryCode } from './catalog';

export const circuitFacts: Record<string, { length: string; laps: number; distance: string; lapRecord: string }> = {
  'Circuit of the Americas': { length: '5,513 km', laps: 56, distance: '308,405 km', lapRecord: '1:36.169 (C. Leclerc, 2019)' },
};

export type Sanction = {
  id: number;
  category: CategoryCode;
  teamId: number;
  subject: string;
  penalty: string;
  event: string;
  acknowledged: boolean;
};

export const sanctions: Sanction[] = [
  { id: 1, category: 'f1', teamId: 2, subject: 'Max Verstappen', penalty: '5 s por exceder límites de pista', event: 'GP de Singapur', acknowledged: true },
  { id: 2, category: 'f1', teamId: 1, subject: 'Lewis Hamilton', penalty: 'Reprimenda por impeding en clasificación', event: 'GP de Singapur', acknowledged: false },
  { id: 3, category: 'f1', teamId: 3, subject: 'Mercedes-AMG Petronas', penalty: 'Multa de 10.000 € por liberación insegura', event: 'GP de Singapur', acknowledged: false },
  { id: 4, category: 'f2', teamId: 6, subject: 'Andrea Kimi Antonelli', penalty: '3 puestos de grilla', event: 'Ronda 12 — Bakú', acknowledged: true },
];

export type ScoreNotice = { id: number; teamId: number; event: string; points: number; acknowledged: boolean };

export const scoreNotices: ScoreNotice[] = [
  { id: 1, teamId: 1, event: 'GP de Singapur', points: 33, acknowledged: false },
  { id: 2, teamId: 2, event: 'GP de Singapur', points: 25, acknowledged: true },
  { id: 3, teamId: 3, event: 'GP de Singapur', points: 18, acknowledged: false },
];
