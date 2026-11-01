import { UnreadMessagesProvider, useUnreadMessages } from '@/data/UnreadMessagesProvider';
import { RoleGate } from '@/navigation/RoleGate';
import { RoleTabs } from '@/navigation/RoleTabs';
import { useFiaTabBadges } from '@/navigation/tabBadges';
import { fiaTabs } from '@/navigation/tabSpecs';

export default function FiaLayout() {
  return (
    <RoleGate roles={['fia_admin']}>
      <UnreadMessagesProvider>
        <FiaTabs />
      </UnreadMessagesProvider>
    </RoleGate>
  );
}

function FiaTabs() {
  const badges = useFiaTabBadges();
  const { reload: reloadUnread } = useUnreadMessages();

  // Reloading on every tab focus keeps the badge current without polling: a
  // message a team sends shows up on the next tab.
  return <RoleTabs tabs={fiaTabs} badges={badges} onTabFocus={() => void reloadUnread()} />;
}
