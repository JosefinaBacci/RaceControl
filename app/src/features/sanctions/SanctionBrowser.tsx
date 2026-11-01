import type { Sanction, SanctionQuery } from '@/api/sanctions';
import { useSanctions } from '@/data/sanctions';
import { RemoteContent } from '@/features/RemoteContent';
import { SanctionList } from '@/features/SanctionList';

import { sanctionSubjects } from './sanctionFilter';
import { SanctionFilters } from './SanctionFilters';
import { useSanctionFilter } from './useSanctionFilter';

type BrowserProps = { onOpen: (sanction: Sanction) => void };

function Results({ sanctions, onOpen }: BrowserProps & { sanctions: Sanction[] }) {
  return <SanctionList title="Sanciones" sanctions={sanctions} onOpen={onOpen} emptyText="No hay sanciones que coincidan con los filtros." />;
}

function FilteredResults({ query, onOpen }: BrowserProps & { query: SanctionQuery }) {
  const sanctions = useSanctions(query);
  return (
    <RemoteContent state={sanctions.state} onRetry={sanctions.reload}>
      {(list) => <Results sanctions={list} onOpen={onOpen} />}
    </RemoteContent>
  );
}

export function SanctionBrowser({ onOpen }: BrowserProps) {
  const history = useSanctions({});
  const filter = useSanctionFilter();
  const filtered = Object.values(filter.query).some((value) => value !== undefined);

  return (
    <RemoteContent state={history.state} onRetry={history.reload}>
      {(list) => (
        <>
          <SanctionFilters value={filter.value} onChange={filter.setValue} errors={filter.errors} {...sanctionSubjects(list)} />
          {filtered ? <FilteredResults query={filter.query} onOpen={onOpen} /> : <Results sanctions={list} onOpen={onOpen} />}
        </>
      )}
    </RemoteContent>
  );
}
