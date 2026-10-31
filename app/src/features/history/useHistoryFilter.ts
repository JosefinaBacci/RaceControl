import { useMemo, useRef, useState } from 'react';

import type { HistoryQuery } from '@/api/history';
import { useDebouncedValue } from '@/data/useDebouncedValue';

import { emptyHistoryFilter, historyQuery, validateDateRange, type HistoryFilterValue } from './historyFilter';

const typingPauseMs = 400;

// Holds what the viewer types and hands the list only a query it can send: while
// a bound is half typed or invalid, the last valid query stays in place.
export function useHistoryFilter() {
  const [value, setValue] = useState<HistoryFilterValue>(emptyHistoryFilter);
  const debounced = useDebouncedValue(value, typingPauseMs);
  const lastValidQuery = useRef<HistoryQuery>(historyQuery(emptyHistoryFilter));

  const query = useMemo(() => {
    const errors = validateDateRange(debounced);
    if (!errors.from && !errors.to) {
      lastValidQuery.current = historyQuery(debounced);
    }
    return lastValidQuery.current;
  }, [debounced]);

  return { value, setValue, errors: validateDateRange(value), query };
}
