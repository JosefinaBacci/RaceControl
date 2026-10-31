// History days are whole UTC days (the API filters and records them that way),
// so they are formatted from their digits and never through the device's zone:
// a viewer at UTC-3 must see the same day the filter matched.

const dayPattern = /^(\d{4})-(\d{2})-(\d{2})$/;

export function isDay(value: string): boolean {
  const match = dayPattern.exec(value);
  if (!match) {
    return false;
  }
  const [year, month, day] = match.slice(1).map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day;
}

export function formatDay(day: string): string {
  const match = dayPattern.exec(day);
  return match ? `${match[3]}/${match[2]}/${match[1]}` : day;
}

export function dayOfInstant(instant: string | Date): string {
  return new Date(instant).toISOString().slice(0, 10);
}

export function formatInstantDay(instant: string): string {
  return formatDay(dayOfInstant(instant));
}

export function todayDay(now: Date = new Date()): string {
  return dayOfInstant(now);
}
