import { useState } from 'react';

import { AppText, ChipGroup, Screen } from '@/components';
import { useSanctionFollowUps } from '@/data/sanctions';
import { useScoreChanges } from '@/data/scores';
import { RemoteContent } from '@/features/RemoteContent';
import { openFiaSanction } from '@/features/sanctions/navigation';
import { SanctionFollowUpCard } from '@/features/sanctions/SanctionFollowUpCard';
import { ScoreChangeCard } from '@/features/scores/ScoreChangeCard';

type Topic = 'scores' | 'sanctions';

const topicOptions: readonly { value: Topic; label: string }[] = [
  { value: 'scores', label: 'Puntajes' },
  { value: 'sanctions', label: 'Sanciones' },
];

export default function FiaNotificationsScreen() {
  const [topic, setTopic] = useState<Topic>('scores');

  return (
    <Screen title="Acuses" subtitle="Qué escuderías confirmaron cada publicación de puntajes y cada sanción.">
      <ChipGroup options={topicOptions} selected={topic} onSelect={setTopic} accessibilityLabel="Elegir acuses" variant="underline" />
      {topic === 'scores' ? <ScoreAcknowledgements /> : <SanctionAcknowledgements />}
    </Screen>
  );
}

function ScoreAcknowledgements() {
  const { state, reload } = useScoreChanges();

  return (
    <RemoteContent state={state} onRetry={reload}>
      {(changes) =>
        changes.length === 0 ? (
          <AppText color="textMuted">Todavía no se publicaron puntajes.</AppText>
        ) : (
          changes.map((change) => <ScoreChangeCard key={change.id} change={change} />)
        )
      }
    </RemoteContent>
  );
}

function SanctionAcknowledgements() {
  const { state, reload } = useSanctionFollowUps();

  return (
    <RemoteContent state={state} onRetry={reload}>
      {(followUps) =>
        followUps.length === 0 ? (
          <AppText color="textMuted">Todavía no se registraron sanciones.</AppText>
        ) : (
          followUps.map((followUp) => (
            <SanctionFollowUpCard key={followUp.sanction.id} followUp={followUp} onOpen={() => openFiaSanction(followUp.sanction)} />
          ))
        )
      }
    </RemoteContent>
  );
}
