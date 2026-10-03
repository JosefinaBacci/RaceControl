import type { ImageSourcePropType } from 'react-native';

import type { Role } from '@/api/auth';

const fiaLogo: ImageSourcePropType = require('../assets/logos/fia.png');

// Keyed by the fixed team ids of backend/seeds/reference.sql.
const teamLogos: Record<number, ImageSourcePropType> = {
  1: require('../assets/logos/ferrari.png'),
  2: require('../assets/logos/red_bull.png'),
  3: require('../assets/logos/mercedes.png'),
};

// The demo accounts of backend/internal/seed/demo.go represent their organization.
const demoAccountAvatars: Record<string, ImageSourcePropType> = {
  'fia.admin': fiaLogo,
  'ferrari.admin': teamLogos[1],
  'redbull.admin': teamLogos[2],
  'mercedes.admin': teamLogos[3],
};

export function avatarImage(username: string): ImageSourcePropType | undefined {
  return demoAccountAvatars[username];
}

export function teamLogo(teamId: number | null): ImageSourcePropType | undefined {
  return teamId === null ? undefined : teamLogos[teamId];
}

export function accountLogo(role: Role, teamId: number | null): ImageSourcePropType | undefined {
  return role === 'fia_admin' ? fiaLogo : teamLogo(teamId);
}
