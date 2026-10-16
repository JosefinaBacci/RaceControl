import type { ScoreEntryInput, ScoreGrid } from '@/api/scores';

// A driver's line is either the points typed for them or the explicit mark that
// they did not race. A blank field is neither, so it never silently drops a score.
export type ScoreLine = { didRace: true; points: string } | { didRace: false };

export type ScoreSheet = Record<number, ScoreLine>;

export const missingPointsMessage = 'Completá los puntos o marcá que no corrió.';
export const invalidPointsMessage = 'Tiene que ser un número mayor o igual a cero.';

// A race with results keeps its unscored drivers out; a fresh race asks for every
// driver, so nobody is left out by omission.
export function initialSheet(grid: ScoreGrid): ScoreSheet {
  const hasResults = grid.entries.some((entry) => entry.points !== null);
  const sheet: ScoreSheet = {};
  for (const entry of grid.entries) {
    sheet[entry.driverId] = initialLine(entry.points, hasResults);
  }
  return sheet;
}

function initialLine(points: number | null, raceHasResults: boolean): ScoreLine {
  if (points !== null) {
    return { didRace: true, points: String(points) };
  }
  return raceHasResults ? { didRace: false } : { didRace: true, points: '' };
}

export function lineError(line: ScoreLine | undefined): string | null {
  if (!line?.didRace) {
    return null;
  }
  const raw = line.points.trim();
  if (raw === '') {
    return missingPointsMessage;
  }
  const value = Number(raw);
  return Number.isFinite(value) && value >= 0 ? null : invalidPointsMessage;
}

// The entries the API takes, or null while any line is blank or invalid.
export function sheetEntries(grid: ScoreGrid, sheet: ScoreSheet): ScoreEntryInput[] | null {
  const entries: ScoreEntryInput[] = [];
  for (const { driverId } of grid.entries) {
    const line = sheet[driverId];
    if (lineError(line) !== null) {
      return null;
    }
    if (line?.didRace) {
      entries.push({ driverId, points: Number(line.points.trim()) });
    }
  }
  return entries;
}
