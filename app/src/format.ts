const dateFormatter = new Intl.DateTimeFormat('es-AR', { weekday: 'short', day: 'numeric', month: 'short' });

export function formatEventDate(isoDate: string): string {
  const [year, month, day] = isoDate.slice(0, 10).split('-').map(Number);
  return dateFormatter.format(new Date(year, month - 1, day));
}

const dateTimeFormatter = new Intl.DateTimeFormat('es-AR', {
  day: 'numeric',
  month: 'short',
  hour: '2-digit',
  minute: '2-digit',
});

export function formatDateTime(isoTimestamp: string): string {
  return dateTimeFormatter.format(new Date(isoTimestamp));
}

const pointsFormatter = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 2 });

export function formatPoints(points: number): string {
  return pointsFormatter.format(points);
}
