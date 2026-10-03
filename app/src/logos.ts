import type { ImageSourcePropType } from 'react-native';

import type { Role } from '@/api/auth';

const fiaLogo: ImageSourcePropType = require('../assets/logos/fia.png');

// Keyed by the fixed team ids of backend/seeds/reference.sql.
const teamLogos: Record<number, ImageSourcePropType> = {
  1: require('../assets/logos/ferrari.png'),
  2: require('../assets/logos/red_bull.png'),
  3: require('../assets/logos/mercedes.png'),
};

export function teamLogo(teamId: number | null): ImageSourcePropType | undefined {
  return teamId === null ? undefined : teamLogos[teamId];
}

export function accountLogo(role: Role, teamId: number | null): ImageSourcePropType | undefined {
  return role === 'fia_admin' ? fiaLogo : teamLogo(teamId);
}
