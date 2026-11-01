import { AppText, Screen } from '@/components';
import { useTeamSanctionNotices } from '@/data/TeamSanctionNoticesProvider';
import { RemoteContent } from '@/features/RemoteContent';
import { SanctionNoticeCard } from '@/features/sanctions/SanctionNoticeCard';

export default function TeamSanctionInboxScreen() {
  const { state, reload } = useTeamSanctionNotices();

  return (
    <Screen
      title="Sanciones"
      subtitle="Sanciones registradas o modificadas por la FIA en tu campeonato. Confirmá a la FIA que las recibiste."
    >
      <RemoteContent state={state} onRetry={reload}>
        {(board) =>
          board.notices.length === 0 ? (
            <AppText color="textMuted">La FIA todavía no registró sanciones en tu campeonato.</AppText>
          ) : (
            board.notices.map((notice) => <SanctionNoticeCard key={notice.id} notice={notice} />)
          )
        }
      </RemoteContent>
    </Screen>
  );
}
