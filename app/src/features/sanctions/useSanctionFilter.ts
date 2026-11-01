import { useMemo, useRef, useState } from 'react';

import type { SanctionQuery } from '@/api/sanctions';
import { useDebouncedValue } from '@/data/useDebouncedValue';
import { validateDateRange } from '@/features/history/historyFilter';

import { emptySanctionFilter, sanctionQuery, type SanctionFilterValue } from './sanctionFilter';

const typingPauseMs = 400;

// Like useHistoryFilter: while a date is half typed or invalid, the last valid
// query stays in place, so the list never asks for a range the API refuses.
export function useSanctionFilter() {
  const [value, setValue] = useState<SanctionFilterValue>(emptySanctionFilter);
  const debounced = useDebouncedValue(value, typingPauseMs);
  const lastValidQuery = useRef<SanctionQuery>(sanctionQuery(emptySanctionFilter));

  const query = useMemo(() => {
    const errors = validateDateRange(debounced);
    if (!errors.from && !errors.to) {
      lastValidQuery.current = sanctionQuery(debounced);
    }
    return lastValidQuery.current;
  }, [debounced]);

  return { value, setValue, errors: validateDateRange(value), query };
}
