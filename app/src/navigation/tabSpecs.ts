import type Ionicons from '@expo/vector-icons/Ionicons';
import type { Href } from 'expo-router';
import type { ComponentProps } from 'react';

export type TabIconName = ComponentProps<typeof Ionicons>['name'];

export type TabSpec = {
  name: string;
  title: string;
  icon: TabIconName;
  href: Href;
};

export type TabBadges = Partial<Record<string, number>>;

export const publicTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Inicio', icon: 'home-outline', href: '/' },
  { name: 'calendar', title: 'Carreras', icon: 'flag-outline', href: '/calendar' },
  { name: 'drivers', title: 'Pilotos', icon: 'people-outline', href: '/drivers' },
  { name: 'standings', title: 'Puntajes', icon: 'trophy-outline', href: '/standings' },
  { name: 'sanctions', title: 'Sanciones', icon: 'alert-circle-outline', href: '/sanctions' },
  { name: 'technical', title: 'Técnica', icon: 'construct-outline', href: '/technical' },
];

export const teamTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Mi escudería', icon: 'speedometer-outline', href: '/team' },
  { name: 'drivers', title: 'Pilotos', icon: 'person-outline', href: '/team/drivers' },
  { name: 'scores', title: 'Puntajes', icon: 'trophy-outline', href: '/team/scores' },
  { name: 'inbox', title: 'Sanciones', icon: 'alert-circle-outline', href: '/team/inbox' },
  { name: 'notifications', title: 'Notificaciones', icon: 'notifications-outline', href: '/team/notifications' },
  { name: 'messages', title: 'Mensajes', icon: 'chatbubbles-outline', href: '/team/messages' },
  { name: 'public', title: 'Sitio público', icon: 'globe-outline', href: '/team/public' },
  { name: 'assignments', title: 'Asignaciones', icon: 'time-outline', href: '/team/assignments' },
];

export const fiaTabs: readonly TabSpec[] = [
  { name: 'index', title: 'Panel', icon: 'speedometer-outline', href: '/fia' },
  { name: 'users', title: 'Usuarios', icon: 'people-outline', href: '/fia/users' },
  { name: 'calendar', title: 'Calendario', icon: 'calendar-outline', href: '/fia/calendar' },
  { name: 'results', title: 'Resultados', icon: 'trophy-outline', href: '/fia/results' },
  { name: 'sanctions', title: 'Sanciones', icon: 'alert-circle-outline', href: '/fia/sanctions' },
  { name: 'notifications', title: 'Acuses', icon: 'checkmark-done-outline', href: '/fia/notifications' },
  { name: 'messages', title: 'Mensajes', icon: 'chatbubbles-outline', href: '/fia/messages' },
  { name: 'public', title: 'Sitio público', icon: 'globe-outline', href: '/fia/public' },
  { name: 'assignments', title: 'Asignaciones', icon: 'time-outline', href: '/fia/assignments' },
];

