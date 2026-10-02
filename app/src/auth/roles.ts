import type { Role } from '@/api/auth';

export const roleHome = {
  fia_admin: '/fia',
  team_admin: '/team',
} as const satisfies Record<Role, string>;

export const roleLabel: Record<Role, string> = {
  fia_admin: 'Administrador FIA',
  team_admin: 'Administrador de escudería',
};
