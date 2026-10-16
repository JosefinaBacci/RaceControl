import { AppText, Screen } from '@/components';
import { useScoreChanges } from '@/data/scores';
import { RemoteContent } from '@/features/RemoteContent';
import { ScoreChangeCard } from '@/features/scores/ScoreChangeCard';

export default function FiaNotificationsScreen() {
  const { state, reload } = useScoreChanges();

  return (
    <Screen title="Acuses de puntajes" subtitle="Qué escuderías confirmaron cada publicación o corrección de puntajes.">
      <RemoteContent state={state} onRetry={reload}>
        {(changes) =>
          changes.length === 0 ? (
            <AppText color="textMuted">Todavía no se publicaron puntajes.</AppText>
          ) : (
            changes.map((change) => <ScoreChangeCard key={change.id} change={change} />)
          )
        }
      </RemoteContent>
    </Screen>
  );
}
