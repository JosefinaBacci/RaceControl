import { useSession } from '@/auth/SessionProvider';
import { AppText, Screen } from '@/components';
import { useTeam } from '@/data/teams';
import { RemoteContent } from '@/features/RemoteContent';
import { ScoreBoard } from '@/features/scores/ScoreBoard';

export default function TeamScoresScreen() {
  const { state, reload } = useTeam(useSession().user?.teamId ?? null);

  return (
    <Screen title="Puntajes" subtitle="Resultados publicados por la FIA y clasificación del campeonato.">
      <RemoteContent state={state} onRetry={reload}>
        {(team) =>
          team ? (
            <ScoreBoard initialCategory={team.categoryCode} />
          ) : (
            <AppText color="textMuted">No encontramos tu escudería en el catálogo.</AppText>
          )
        }
      </RemoteContent>
    </Screen>
  );
}
