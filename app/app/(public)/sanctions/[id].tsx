import { router, useLocalSearchParams } from 'expo-router';

import { Button, Screen } from '@/components';
import { useSanction } from '@/data/sanctions';
import { RemoteContent } from '@/features/RemoteContent';
import { SanctionHistoryCard, SanctionSummaryCard } from '@/features/sanctions/SanctionDetail';

export default function SanctionDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>();
  const { state, reload } = useSanction(Number(id));

  return (
    <Screen
      title="Sanción"
      headerAction={<Button label="Volver" icon="arrow-back" variant="ghost" compact onPress={() => router.back()} />}
    >
      <RemoteContent state={state} onRetry={reload}>
        {(sanction) => (
          <>
            <SanctionSummaryCard sanction={sanction} />
            <SanctionHistoryCard sanction={sanction} />
          </>
        )}
      </RemoteContent>
    </Screen>
  );
}
