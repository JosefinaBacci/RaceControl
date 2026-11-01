import type { Sanction } from '@/api/sanctions';
import { useSanctionCatalogs, useSanctions } from '@/data/sanctions';
import { RemoteContent } from '@/features/RemoteContent';
import { SanctionList } from '@/features/SanctionList';

import { SanctionFilters } from './SanctionFilters';
import { useSanctionFilter } from './useSanctionFilter';

// The filtered list the public and the FIA browse; only what opening a
// sanction leads to differs.
export function SanctionBrowser({ onOpen }: { onOpen: (sanction: Sanction) => void }) {
  const catalogs = useSanctionCatalogs();
  const filter = useSanctionFilter();
  const sanctions = useSanctions(filter.query);

  return (
    <>
      <RemoteContent state={catalogs.state} onRetry={catalogs.reload}>
        {({ drivers, teams }) => (
          <SanctionFilters value={filter.value} onChange={filter.setValue} errors={filter.errors} drivers={drivers} teams={teams} />
        )}
      </RemoteContent>
      <RemoteContent state={sanctions.state} onRetry={sanctions.reload}>
        {(list) => (
          <SanctionList title="Sanciones" sanctions={list} onOpen={onOpen} emptyText="No hay sanciones que coincidan con los filtros." />
        )}
      </RemoteContent>
    </>
  );
}
