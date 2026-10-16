import { Screen } from '@/components';
import { ScoreBoard } from '@/features/scores/ScoreBoard';

export default function PublicStandingsScreen() {
  return (
    <Screen title="Puntajes" subtitle="Resultados de cada carrera y clasificación del campeonato por categoría.">
      <ScoreBoard initialCategory="f1" />
    </Screen>
  );
}
