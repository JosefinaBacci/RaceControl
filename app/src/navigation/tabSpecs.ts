import type Ionicons from '@expo/vector-icons/Ionicons';
import type { Href } from 'expo-router';
import type { ComponentProps } from 'react';

export type TabIconName = ComponentProps<typeof Ionicons>['name'];

export type TabSpec = {
  name: string;
  title: string;
  icon: TabIconName;
  href: Href;
  isPhonePrimary?: boolean;
  phoneLabel?: string;
};

export type TabBadges = Partial<Record<string, number>>;

// The phone tab bar fits five labels of up to ten characters at 320 px; the "Más"
// tab takes the fifth slot.
export const maxPhonePrimaryTabs = 4;
export const maxPhoneLabelLength = 10;

export const moreTab = { name: 'more', title: 'Más', icon: 'ellipsis-horizontal' } as const satisfies Omit<TabSpec, 'href'>;

export const publicTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Inicio', icon: 'home-outline', href: '/', isPhonePrimary: true },
  { name: 'calendar', title: 'Carreras', icon: 'flag-outline', href: '/calendar', isPhonePrimary: true },
  { name: 'drivers', title: 'Pilotos', icon: 'people-outline', href: '/drivers' },
  { name: 'standings', title: 'Puntajes', icon: 'trophy-outline', href: '/standings', isPhonePrimary: true },
  { name: 'sanctions', title: 'Sanciones', icon: 'alert-circle-outline', href: '/sanctions', isPhonePrimary: true },
  { name: 'technical', title: 'Técnica', icon: 'construct-outline', href: '/technical' },
];

export const teamTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Mi escudería', icon: 'speedometer-outline', href: '/team', isPhonePrimary: true, phoneLabel: 'Escudería' },
  { name: 'drivers', title: 'Pilotos', icon: 'person-outline', href: '/team/drivers', isPhonePrimary: true },
  { name: 'scores', title: 'Puntajes', icon: 'trophy-outline', href: '/team/scores' },
  { name: 'inbox', title: 'Sanciones', icon: 'alert-circle-outline', href: '/team/inbox', isPhonePrimary: true },
  { name: 'notifications', title: 'Notificaciones', icon: 'notifications-outline', href: '/team/notifications' },
  { name: 'messages', title: 'Mensajes', icon: 'chatbubbles-outline', href: '/team/messages', isPhonePrimary: true },
  { name: 'public', title: 'Sitio público', icon: 'globe-outline', href: '/team/public' },
  { name: 'assignments', title: 'Asignaciones', icon: 'time-outline', href: '/team/assignments' },
];

export const fiaTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Panel', icon: 'speedometer-outline', href: '/fia', isPhonePrimary: true },
  { name: 'users', title: 'Usuarios', icon: 'people-outline', href: '/fia/users' },
  { name: 'calendar', title: 'Calendario', icon: 'calendar-outline', href: '/fia/calendar', isPhonePrimary: true },
  { name: 'results', title: 'Resultados', icon: 'trophy-outline', href: '/fia/results', isPhonePrimary: true },
  { name: 'sanctions', title: 'Sanciones', icon: 'alert-circle-outline', href: '/fia/sanctions', isPhonePrimary: true },
  { name: 'notifications', title: 'Acuses', icon: 'checkmark-done-outline', href: '/fia/notifications' },
  { name: 'messages', title: 'Mensajes', icon: 'chatbubbles-outline', href: '/fia/messages' },
  { name: 'public', title: 'Sitio público', icon: 'globe-outline', href: '/fia/public' },
  { name: 'assignments', title: 'Asignaciones', icon: 'time-outline', href: '/fia/assignments' },
];

export function phoneTabLabel(tab: TabSpec): string {
  return tab.phoneLabel ?? tab.title;
}

export function moreMenuTabs(tabs: readonly TabSpec[]): TabSpec[] {
  return tabs.filter((tab) => !tab.isPhonePrimary);
}

// "Más" carries the badges of the tabs it hides, so a pending item is never invisible on a phone.
export function moreMenuBadge(tabs: readonly TabSpec[], badges: TabBadges): number {
  return moreMenuTabs(tabs).reduce((total, tab) => total + (badges[tab.name] ?? 0), 0);
}
