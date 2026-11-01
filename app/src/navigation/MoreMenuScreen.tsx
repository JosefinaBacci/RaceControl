import Ionicons from '@expo/vector-icons/Ionicons';
import { router } from 'expo-router';

import { Badge, Card, ListItem, Screen } from '@/components';
import { colors } from '@/theme';

import { moreMenuTabs, type TabBadges, type TabSpec } from './tabSpecs';

type MoreMenuScreenProps = {
  tabs: readonly TabSpec[];
  badges?: TabBadges;
};

export function MoreMenuScreen({ tabs, badges = {} }: MoreMenuScreenProps) {
  const entries = moreMenuTabs(tabs);

  return (
    <Screen title="Más">
      <Card>
        {entries.map((tab, index) => (
          <MoreMenuEntry key={tab.name} tab={tab} badge={badges[tab.name] ?? 0} isLast={index === entries.length - 1} />
        ))}
      </Card>
    </Screen>
  );
}

function MoreMenuEntry({ tab, badge, isLast }: { tab: TabSpec; badge: number; isLast: boolean }) {
  return (
    <ListItem
      title={tab.title}
      leading={<Ionicons name={tab.icon} size={20} color={colors.text} />}
      trailing={badge > 0 ? <Badge label={String(badge)} tone="accent" /> : undefined}
      isLast={isLast}
      onPress={() => router.navigate(tab.href)}
    />
  );
}
