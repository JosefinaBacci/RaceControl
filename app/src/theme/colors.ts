export const colors = {
  background: '#0B0B0F',
  surface: '#15151B',
  surfaceRaised: '#1D1D25',
  surfaceMuted: '#26262F',
  border: '#2A2A33',
  text: '#F5F5F7',
  textMuted: '#9C9CA8',
  textInverse: '#FFFFFF',
  accent: '#E10600',
  accentPressed: '#B30500',
  accentSoft: 'rgba(225, 6, 0, 0.16)',
  success: '#3DDC84',
  successSoft: 'rgba(61, 220, 132, 0.14)',
  warning: '#FFB020',
  warningSoft: 'rgba(255, 176, 32, 0.14)',
  danger: '#FF5A52',
  dangerSoft: 'rgba(255, 90, 82, 0.14)',
  info: '#5AA9FF',
  infoSoft: 'rgba(90, 169, 255, 0.14)',
  overlay: 'rgba(11, 11, 15, 0.72)',
} as const;

export type ColorName = keyof typeof colors;
