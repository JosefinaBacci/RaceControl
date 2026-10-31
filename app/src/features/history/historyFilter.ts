import { isDay } from './day';

// Days are typed as YYYY-MM-DD; a blank bound is open. Validated here so an
// impossible range never reaches the API, which would refuse it with a 400.

export type DateRangeValue = {
  from: string;
  to: string;
};

export type DateRangeErrors = Partial<Record<keyof DateRangeValue, string>>;

export type HistoryFilterValue = DateRangeValue & {
  type: string;
};

export const allTypes = '';

export const emptyHistoryFilter: HistoryFilterValue = { from: '', to: '', type: allTypes };

const dayFormatError = 'Usá el formato AAAA-MM-DD';

export function validateDateRange(value: DateRangeValue): DateRangeErrors {
  const from = value.from.trim();
  const to = value.to.trim();
  const errors: DateRangeErrors = {};
  if (from && !isDay(from)) {
    errors.from = dayFormatError;
  }
  if (to && !isDay(to)) {
    errors.to = dayFormatError;
  }
  if (!errors.from && !errors.to && from && to && from > to) {
    errors.to = 'La fecha final no puede ser anterior a la inicial';
  }
  return errors;
}

export function dateRangeQuery(value: DateRangeValue): { from?: string; to?: string } {
  return { from: value.from.trim() || undefined, to: value.to.trim() || undefined };
}

export function historyQuery(value: HistoryFilterValue): { from?: string; to?: string; type?: string } {
  return { ...dateRangeQuery(value), type: value.type || undefined };
}
