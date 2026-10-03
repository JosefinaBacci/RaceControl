import type { ImageSourcePropType } from 'react-native';

import type { CategoryCode } from '@/mocks/catalog';

export type Photo = {
  source: ImageSourcePropType;
  credit: string;
};

const f1: Photo = { source: require('../assets/photos/f1.jpg'), credit: 'Foto: Liauzh · CC BY-SA 4.0' };
const f2: Photo = { source: require('../assets/photos/f2.jpg'), credit: 'Foto: Lukas Raich · CC BY-SA 4.0' };
const f3: Photo = { source: require('../assets/photos/f3.jpg'), credit: 'Foto: Lukas Raich · CC BY-SA 4.0' };
const academy: Photo = { source: require('../assets/photos/academy.jpg'), credit: 'Foto: Lukas Raich · CC BY-SA 4.0' };
const heroF1: Photo = { source: require('../assets/photos/hero-f1.jpg'), credit: 'Foto: Liauzh · CC BY-SA 4.0' };
const loginFerrari: Photo = { source: require('../assets/photos/login-ferrari.webp'), credit: 'Imagen: Scuderia Ferrari (prensa)' };
const teamFerrari: Photo = { source: require('../assets/photos/team-ferrari.jpg'), credit: 'Foto: Lukas Raich · CC BY-SA 4.0' };

export const categoryPhotos: Record<CategoryCode, Photo> = { f1, f2, f3, academy };

const nextRacePhotos: Partial<Record<CategoryCode, Photo>> = { f1: heroF1 };
const teamPhotos: Record<number, Photo> = { 1: teamFerrari };

export const loginPhoto = loginFerrari;

export function nextRacePhoto(category: CategoryCode): Photo {
  return nextRacePhotos[category] ?? categoryPhotos[category];
}

export function teamPhoto(teamId: number, category: CategoryCode): Photo {
  return teamPhotos[teamId] ?? categoryPhotos[category];
}
