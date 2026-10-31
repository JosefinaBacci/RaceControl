import type { ReactNode } from 'react';
import { StyleSheet, View } from 'react-native';

import type { Sanction } from '@/api/sanctions';
import { AppText, Badge, Card, ListItem } from '@/components';
import { formatInstantDay } from '@/features/history/day';
import { actionLabels, penaltyText, subjectText } from '@/features/sanctions/sanctionText';
import { teamById } from '@/mocks/catalog';

type SanctionListProps = {
  title: string;
  sanctions: Sanction[];
  action?: ReactNode;
  onOpen?: (sanction: Sanction) => void;
  emptyText?: string;
};

export function SanctionList({ title, sanctions, action, onOpen, emptyText = 'No hay sanciones registradas.' }: SanctionListProps) {
  return (
    <Card title={title} action={action}>
      {sanctions.length === 0 ? <AppText color="textMuted">{emptyText}</AppText> : null}
      {sanctions.map((sanction, index) => (
        <ListItem
          key={sanction.id}
          leading={<View style={[styles.marker, { backgroundColor: teamById(sanction.team.id).color }]} />}
          title={penaltyText(sanction.current.penalty)}
          subtitle={sanctionSummary(sanction)}
          trailing={<Badge label={actionLabels[sanction.current.action].label} tone={actionLabels[sanction.current.action].tone} />}
          onPress={onOpen ? () => onOpen(sanction) : undefined}
          isLast={index === sanctions.length - 1}
        />
      ))}
    </Card>
  );
}

function sanctionSummary(sanction: Sanction): string {
  const parts = [subjectText(sanction.subject, sanction.team), sanction.current.event?.name, formatInstantDay(sanction.issuedAt)];
  return parts.filter(Boolean).join(' · ');
}

const styles = StyleSheet.create({
  marker: { width: 4, alignSelf: 'stretch', borderRadius: 2 },
});
