import { RoleTabs } from '@/navigation/RoleTabs';
import { publicTabs } from '@/navigation/tabSpecs';

export default function PublicLayout() {
  return <RoleTabs tabs={publicTabs} />;
}
