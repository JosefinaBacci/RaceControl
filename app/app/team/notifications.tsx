import { AppText, Screen } from '@/components';
import { useReloadOnFocus } from '@/data/useReloadOnFocus';
import { useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { RemoteContent } from '@/features/RemoteContent';
import { ScoreNoticeCard } from '@/features/scores/ScoreNoticeCard';

export default function TeamNotificationsScreen() {
  const { state, reload } = useTeamScoreNotices();
  useReloadOnFocus(reload);

  return (
    <Screen title="Notificaciones" subtitle="Publicaciones y correcciones de puntajes de tu escudería. Confirmá a la FIA que las recibiste.">
      <RemoteContent state={state} onRetry={reload}>
        {(board) =>
          board.notices.length === 0 ? (
            <AppText color="textMuted">Todavía no hay puntajes publicados para tu escudería.</AppText>
          ) : (
            board.notices.map((notice) => <ScoreNoticeCard key={notice.id} notice={notice} />)
          )
        }
      </RemoteContent>
    </Screen>
  );
}
