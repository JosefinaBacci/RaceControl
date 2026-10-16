import { useSession } from '@/auth/SessionProvider';
import { Screen } from '@/components';
import { ScoreBoard } from '@/features/scores/ScoreBoard';
import { teamById } from '@/mocks/catalog';

export default function TeamScoresScreen() {
  const team = teamById(useSession().user?.teamId ?? null);

  return (
    <Screen title="Puntajes" subtitle="Resultados publicados por la FIA y clasificación del campeonato.">
      <ScoreBoard initialCategory={team.category} />
    </Screen>
  );
}
