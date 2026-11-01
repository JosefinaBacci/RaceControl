import { TeamSanctionNoticesProvider, useTeamSanctionNotices } from '@/data/TeamSanctionNoticesProvider';
import { TeamScoreNoticesProvider, useTeamScoreNotices } from '@/data/TeamScoreNoticesProvider';
import { UnreadMessagesProvider, useUnreadMessages } from '@/data/UnreadMessagesProvider';
import { RoleGate } from '@/navigation/RoleGate';
import { RoleTabs } from '@/navigation/RoleTabs';
import { useTeamTabBadges } from '@/navigation/tabBadges';
import { teamTabs } from '@/navigation/tabSpecs';

export default function TeamLayout() {
  return (
    <RoleGate roles={['team_admin']}>
      <TeamScoreNoticesProvider>
        <TeamSanctionNoticesProvider>
          <UnreadMessagesProvider>
            <TeamTabs />
          </UnreadMessagesProvider>
        </TeamSanctionNoticesProvider>
      </TeamScoreNoticesProvider>
    </RoleGate>
  );
}

function TeamTabs() {
  const badges = useTeamTabBadges();
  const { reload: reloadNotices } = useTeamScoreNotices();
  const { reload: reloadSanctionNotices } = useTeamSanctionNotices();
  const { reload: reloadUnread } = useUnreadMessages();

  // Reloading on every tab focus keeps the badges current wherever the team admin
  // is, without polling: a change the FIA publishes or a message it sends shows
  // up on the next tab.
  const reloadBadges = () => {
    void reloadNotices();
    void reloadSanctionNotices();
    void reloadUnread();
  };

  return <RoleTabs tabs={teamTabs} badges={badges} onTabFocus={reloadBadges} />;
}
