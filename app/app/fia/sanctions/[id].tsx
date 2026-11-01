import { router, useLocalSearchParams } from 'expo-router';

import type { SanctionDetail } from '@/api/sanctions';
import { Button, Screen } from '@/components';
import { useSanction, useSanctionCatalogs, useSanctionFollowUps } from '@/data/sanctions';
import { RemoteContent } from '@/features/RemoteContent';
import { AmendSanctionCard, RevokeSanctionCard } from '@/features/sanctions/SanctionAdminForms';
import { SanctionHistoryCard, SanctionSummaryCard } from '@/features/sanctions/SanctionDetail';
import { SanctionFollowUpCard } from '@/features/sanctions/SanctionFollowUpCard';

export default function FiaSanctionScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const sanction = useSanction(Number(id));
  const followUps = useSanctionFollowUps();

  const reloadAll = () => {
    void sanction.reload();
    void followUps.reload();
  };

  return (
    <Screen
      title="Sanción"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={sanction.state} onRetry={sanction.reload}>
        {(detail) => (
          <>
            <SanctionSummaryCard sanction={detail} />
            <SanctionHistoryCard sanction={detail} />
            {detail.status === 'active' ? <SanctionChanges key={detail.current.id} sanction={detail} onSaved={reloadAll} /> : null}
          </>
        )}
      </RemoteContent>
      <RemoteContent state={followUps.state} onRetry={followUps.reload}>
        {(list) => {
          const followUp = list.find((candidate) => candidate.sanction.id === Number(id));
          return followUp ? <SanctionFollowUpCard followUp={followUp} /> : null;
        }}
      </RemoteContent>
    </Screen>
  );
}

// Keyed by the standing version, so the forms start over from the terms that
// stand after every change.
function SanctionChanges({ sanction, onSaved }: { sanction: SanctionDetail; onSaved: () => void }) {
  const catalogs = useSanctionCatalogs();

  return (
    <RemoteContent state={catalogs.state} onRetry={catalogs.reload}>
      {({ penaltyTypes }) => (
        <>
          <AmendSanctionCard sanction={sanction} penaltyTypes={penaltyTypes} onSaved={onSaved} />
          <RevokeSanctionCard sanction={sanction} onSaved={onSaved} />
        </>
      )}
    </RemoteContent>
  );
}
