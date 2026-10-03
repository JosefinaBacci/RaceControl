import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import { Card, ListItem } from '@/components';
import { teamById } from '@/mocks/catalog';
import type { Sanction } from '@/mocks/data';

type SanctionListProps = {
  title: string;
  sanctions: Sanction[];
  trailing?: (sanction: Sanction) => ReactNode;
  action?: ReactNode;
};

export function SanctionList({ title, sanctions, trailing, action }: SanctionListProps) {
  return (
    <Card title={title} action={action}>
      {sanctions.map((sanction, index) => (
        <ListItem
          key={sanction.id}
          leading={<View style={[styles.marker, { backgroundColor: teamById(sanction.teamId).color }]} />}
          title={sanction.penalty}
          subtitle={`${sanction.subject} · ${sanction.event}`}
          trailing={trailing?.(sanction)}
          isLast={index === sanctions.length - 1}
        />
      ))}
    </Card>
  );
}

const styles = StyleSheet.create({
  marker: { width: 4, alignSelf: 'stretch', borderRadius: 2 },
});
