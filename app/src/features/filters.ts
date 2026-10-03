import { categories } from '@/mocks/catalog';

export const categoryFilterOptions = [{ value: 'all' as const, label: 'Todas' }, ...categories];

export const periodOptions = [
  { value: 'upcoming' as const, label: 'Próximas' },
  { value: 'past' as const, label: 'Pasadas' },
];
