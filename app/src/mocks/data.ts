import type { FlagCode } from '@/illustrations';

import type { CategoryCode } from './catalog';

export type EventKind = 'race' | 'tyre_test' | 'technical_control';

export type CalendarEvent = {
  id: number;
  category: CategoryCode;
  kind: EventKind;
  name: string;
  circuit: string;
  country: FlagCode;
  date: string;
  startTime?: string;
  status: 'scheduled' | 'finished';
};

export const calendarEvents: CalendarEvent[] = [
  { id: 1, category: 'f1', kind: 'race', name: 'Gran Premio de Singapur', circuit: 'Marina Bay', country: 'sg', date: '2026-10-04', status: 'finished' },
  { id: 2, category: 'f1', kind: 'technical_control', name: 'Control técnico post-carrera', circuit: 'Marina Bay', country: 'sg', date: '2026-10-05', status: 'scheduled' },
  { id: 3, category: 'f1', kind: 'race', name: 'Gran Premio de Estados Unidos', circuit: 'Circuit of the Americas', country: 'us', date: '2026-10-18', startTime: '16:00', status: 'scheduled' },
  { id: 4, category: 'f1', kind: 'tyre_test', name: 'Prueba de neumáticos Pirelli 2027', circuit: 'Circuit of the Americas', country: 'us', date: '2026-10-20', status: 'scheduled' },
  { id: 5, category: 'f2', kind: 'race', name: 'Ronda 13 — Qatar', circuit: 'Lusail', country: 'qa', date: '2026-11-27', status: 'scheduled' },
  { id: 6, category: 'f3', kind: 'race', name: 'Ronda 10 — Monza', circuit: 'Monza', country: 'it', date: '2026-09-06', status: 'finished' },
  { id: 7, category: 'academy', kind: 'race', name: 'Ronda 7 — Las Vegas', circuit: 'Las Vegas Strip', country: 'us', date: '2026-11-21', status: 'scheduled' },
  { id: 8, category: 'f2', kind: 'technical_control', name: 'Verificación de pesos', circuit: 'Lusail', country: 'qa', date: '2026-11-26', status: 'scheduled' },
];

export const eventKindLabel: Record<EventKind, string> = {
  race: 'Carrera',
  tyre_test: 'Prueba de neumáticos',
  technical_control: 'Control técnico',
};

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

export type Driver = { id: number; teamId: number; name: string; number: number; status: 'starter' | 'reserve' };

export const drivers: Driver[] = [
  { id: 1, teamId: 1, name: 'Charles Leclerc', number: 16, status: 'starter' },
  { id: 2, teamId: 1, name: 'Lewis Hamilton', number: 44, status: 'starter' },
  { id: 3, teamId: 1, name: 'Antonio Giovinazzi', number: 99, status: 'reserve' },
  { id: 4, teamId: 2, name: 'Max Verstappen', number: 1, status: 'starter' },
  { id: 5, teamId: 2, name: 'Yuki Tsunoda', number: 22, status: 'starter' },
  { id: 6, teamId: 3, name: 'George Russell', number: 63, status: 'starter' },
  { id: 7, teamId: 3, name: 'Andrea Kimi Antonelli', number: 12, status: 'starter' },
];

export type Notification = { id: number; title: string; body: string; sentAt: string; acknowledgedBy: number; recipients: number };

export const notifications: Notification[] = [
  { id: 1, title: 'Nuevo reglamento técnico 2027', body: 'Disponible para descarga la versión 1.2 del reglamento técnico.', sentAt: '2026-09-30', acknowledgedBy: 7, recipients: 10 },
  { id: 2, title: 'Cambio de horario — GP de Estados Unidos', body: 'La FP1 se adelanta 30 minutos.', sentAt: '2026-10-01', acknowledgedBy: 4, recipients: 10 },
];

export type ManagedUser = { id: number; username: string; email: string; role: 'fia_admin' | 'team_admin'; teamId: number | null; isActive: boolean };

export const managedUsers: ManagedUser[] = [
  { id: 1, username: 'fia.admin', email: 'fia.admin@racecontrol.test', role: 'fia_admin', teamId: null, isActive: true },
  { id: 2, username: 'ferrari.admin', email: 'ferrari.admin@racecontrol.test', role: 'team_admin', teamId: 1, isActive: true },
  { id: 3, username: 'redbull.admin', email: 'redbull.admin@racecontrol.test', role: 'team_admin', teamId: 2, isActive: true },
  { id: 4, username: 'mercedes.admin', email: 'mercedes.admin@racecontrol.test', role: 'team_admin', teamId: 3, isActive: true },
  { id: 5, username: 'mclaren.admin', email: 'mclaren.admin@racecontrol.test', role: 'team_admin', teamId: 4, isActive: false },
];
