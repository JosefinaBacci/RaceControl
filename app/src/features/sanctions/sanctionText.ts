import type { Penalty, PenaltyUnit, SanctionAction, SanctionSubject, SanctionTeam } from '@/api/sanctions';
import type { BadgeTone } from '@/components';

// How each unit reads after its value; the catalogue itself lives in the API.
const unitSuffixes: Record<PenaltyUnit, { one: string; many: string }> = {
  seconds: { one: 's', many: 's' },
  grid_places: { one: 'puesto', many: 'puestos' },
  license_points: { one: 'punto', many: 'puntos' },
  euros: { one: '€', many: '€' },
  races: { one: 'carrera', many: 'carreras' },
};

const thousands = new Intl.NumberFormat('es-AR');

export function penaltyText(penalty: Penalty): string {
  if (penalty.unit === null || penalty.value === null) {
    return penalty.name;
  }
  const suffix = unitSuffixes[penalty.unit];
  return `${penalty.name}: ${thousands.format(penalty.value)} ${penalty.value === 1 ? suffix.one : suffix.many}`;
}

export function unitLabel(unit: PenaltyUnit): string {
  return unitSuffixes[unit].many;
}

export const actionLabels: Record<SanctionAction, { label: string; tone: BadgeTone; verb: string }> = {
  issued: { label: 'Registrada', tone: 'danger', verb: 'Registrada' },
  amended: { label: 'Modificada', tone: 'warning', verb: 'Modificada' },
  revoked: { label: 'Revocada', tone: 'neutral', verb: 'Revocada' },
};

export function subjectText(subject: SanctionSubject, team: SanctionTeam): string {
  if (subject.kind === 'team') {
    return team.name;
  }
  return `${subject.driverFirstName ?? ''} ${subject.driverLastName ?? ''}`.trim() + ` (${team.name})`;
}

// How a team reads a notice: whether the sanction is on itself or on one of
// its drivers, or on another team of the championship.
export function noticeAudienceText(subject: SanctionSubject, isOwnTeam: boolean): string | null {
  if (!isOwnTeam) {
    return null;
  }
  if (subject.kind === 'team') {
    return 'Sanción a tu escudería';
  }
  return `Sanción a tu piloto ${subject.driverFirstName ?? ''} ${subject.driverLastName ?? ''}`.trim();
}

export type SanctionStatusFilter = 'all' | 'active' | 'revoked';

export const statusOptions: readonly { value: SanctionStatusFilter; label: string }[] = [
  { value: 'all', label: 'Todas' },
  { value: 'active', label: 'Vigentes' },
  { value: 'revoked', label: 'Revocadas' },
];
