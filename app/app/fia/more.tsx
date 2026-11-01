import { MoreMenuScreen } from '@/navigation/MoreMenuScreen';
import { useFiaTabBadges } from '@/navigation/tabBadges';
import { fiaTabs } from '@/navigation/tabSpecs';

export default function FiaMoreScreen() {
  return <MoreMenuScreen tabs={fiaTabs} badges={useFiaTabBadges()} />;
}
