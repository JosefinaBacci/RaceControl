import { MoreMenuScreen } from '@/navigation/MoreMenuScreen';
import { useTeamTabBadges } from '@/navigation/tabBadges';
import { teamTabs } from '@/navigation/tabSpecs';

export default function TeamMoreScreen() {
  return <MoreMenuScreen tabs={teamTabs} badges={useTeamTabBadges()} />;
}
