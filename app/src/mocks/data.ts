import type { CategoryCode } from './catalog';

export const circuitFacts: Record<string, { length: string; laps: number; distance: string; lapRecord: string }> = {
  'Circuit of the Americas': { length: '5,513 km', laps: 56, distance: '308,405 km', lapRecord: '1:36.169 (C. Leclerc, 2019)' },
};

export type StandingRow = { position: number; driver: string; teamId: number; points: number };

export const driverStandings: Record<CategoryCode, StandingRow[]> = {
  f1: [
    { position: 1, driver: 'Charles Leclerc', teamId: 1, points: 312 },
    { position: 2, driver: 'Max Verstappen', teamId: 2, points: 298 },
    { position: 3, driver: 'Lando Norris', teamId: 4, points: 287 },
    { position: 4, driver: 'George Russell', teamId: 3, points: 240 },
    { position: 5, driver: 'Lewis Hamilton', teamId: 1, points: 231 },
  ],
  f2: [
    { position: 1, driver: 'Victor Martins', teamId: 5, points: 176 },
    { position: 2, driver: 'Andrea Kimi Antonelli', teamId: 6, points: 168 },
  ],
  f3: [{ position: 1, driver: 'Mari Boya', teamId: 7, points: 121 }],
  academy: [{ position: 1, driver: 'Maya Weug', teamId: 8, points: 98 }],
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

export type Notification = { id: number; title: string; body: string; sentAt: string; acknowledgedBy: number; recipients: number };

export const notifications: Notification[] = [
  { id: 1, title: 'Nuevo reglamento técnico 2027', body: 'Disponible para descarga la versión 1.2 del reglamento técnico.', sentAt: '2026-09-30', acknowledgedBy: 7, recipients: 10 },
  { id: 2, title: 'Cambio de horario — GP de Estados Unidos', body: 'La FP1 se adelanta 30 minutos.', sentAt: '2026-10-01', acknowledgedBy: 4, recipients: 10 },
];
