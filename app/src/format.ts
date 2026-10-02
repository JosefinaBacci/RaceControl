const dateFormatter = new Intl.DateTimeFormat('es-AR', { weekday: 'short', day: 'numeric', month: 'short' });

export function formatEventDate(isoDate: string): string {
  const [year, month, day] = isoDate.split('-').map(Number);
  return dateFormatter.format(new Date(year, month - 1, day));
}
