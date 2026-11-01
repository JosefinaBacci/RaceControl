import { MoreMenuScreen } from '@/navigation/MoreMenuScreen';
import { publicTabs } from '@/navigation/tabSpecs';

export default function PublicMoreScreen() {
  return <MoreMenuScreen tabs={publicTabs} />;
}
